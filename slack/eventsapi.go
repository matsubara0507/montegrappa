package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/matsubara0507/montegrappa/bot"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"net/http"
	"sync"
	"time"
)

type EventApiServer struct {
	mutex     *sync.Mutex
	eventChan chan *bot.Event
	addr      string
	server    *http.Server
	handler   *http.ServeMux

	secretToken   string
	signingSecret string
}

type EventHandlers struct {
	eventMessageHandler func()
}

func NewEventAPIServer(endpoint, addr, secretToken, signingSecret string) *EventApiServer {
	eventApiServer := &EventApiServer{
		mutex:         &sync.Mutex{},
		eventChan:     make(chan *bot.Event),
		addr:          addr,
		secretToken:   secretToken,
		signingSecret: signingSecret,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(endpoint, eventApiServer.receiveEventMessage)

	eventApiServer.handler = mux

	return eventApiServer
}

func (eventApiServer *EventApiServer) Setup() error {
	return nil
}

func (eventApiServer *EventApiServer) Start() error {
	eventApiServer.mutex.Lock()
	if eventApiServer.server == nil {
		eventApiServer.server = &http.Server{
			Addr:         eventApiServer.addr,
			Handler:      eventApiServer.handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 10 * time.Second,
		}
	} else {
		eventApiServer.mutex.Unlock()
		return errors.New("auth server already started")
	}
	eventApiServer.mutex.Unlock()

	return eventApiServer.server.ListenAndServe()
}

func (eventApiServer *EventApiServer) Shutdown(ctx context.Context) error {
	return eventApiServer.server.Shutdown(ctx)
}

func (eventApiServer *EventApiServer) ReceivedEvent() chan *bot.Event {
	return eventApiServer.eventChan
}

func (eventApiServer *EventApiServer) receiveEventMessage(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// ref: https://api.slack.com/authentication/verifying-requests-from-slack
	if err := eventApiServer.secretVerify(r.Header, buf.Bytes()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := buf.String()
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: eventApiServer.secretToken}))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var r *slackevents.ChallengeResponse
		if err := json.Unmarshal([]byte(body), &r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text")
		if _, err := w.Write([]byte(r.Challenge)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		break
	case slackevents.CallbackEvent:
		eventApiServer.eventChan <- eventApiServer.botEvent(eventsAPIEvent.InnerEvent)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (eventApiServer *EventApiServer) secretVerify(header http.Header, body []byte) error {
	verifier, err := slack.NewSecretsVerifier(header, eventApiServer.signingSecret)
	if err != nil {
		return err
	}
	if _, err := verifier.Write(body); err != nil {
		return err
	}
	if err := verifier.Ensure(); err != nil {
		return err
	}
	return nil
}

func (eventApiServer *EventApiServer) botEvent(innerEvent slackevents.EventsAPIInnerEvent) *bot.Event {
	botEvent := new(bot.Event)

	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		botEvent.Type = ev.Type
		botEvent.Message = ev.Text
		botEvent.Channel = ev.Channel
		botEvent.User.Id = ev.User
		botEvent.Ts = ev.TimeStamp
		break
	default:
		return nil
	}

	return botEvent
}
