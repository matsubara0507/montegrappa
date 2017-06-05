package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
	slackOAuth "golang.org/x/oauth2/slack"
)

type AuthServer struct {
	TeamID    string
	TokenChan chan *oauth2.Token
	conf      *oauth2.Config
	server    *http.Server
	mutex     sync.Mutex
}

func NewAuthServer(clientId, secret string, scopes []string, teamId string) *AuthServer {
	authServer := &AuthServer{
		TeamID:    teamId,
		TokenChan: make(chan *oauth2.Token, 1),
		conf: &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: secret,
			Scopes:       scopes,
			Endpoint:     slackOAuth.Endpoint,
		},
	}

	http.HandleFunc("/start", authServer.startApp)
	http.HandleFunc("/callback", authServer.oauthCallback)

	return authServer
}

func (authServer *AuthServer) Start(addr string) error {
	authServer.mutex.Lock()
	if authServer.server == nil {
		s := &http.Server{Addr: addr}
		authServer.server = s
	} else {
		authServer.mutex.Unlock()
		return errors.New("auth server already started")
	}
	authServer.mutex.Unlock()
	return authServer.server.ListenAndServe()
}

func (authServer *AuthServer) startApp(w http.ResponseWriter, r *http.Request) {
	redirectURL := authServer.conf.AuthCodeURL("", oauth2.SetAuthURLParam("team", authServer.TeamID))
	w.Header().Add("Location", redirectURL)
	w.WriteHeader(http.StatusSeeOther)
}

func (authServer *AuthServer) oauthCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	code := r.Form.Get("code")

	token, err := authServer.conf.Exchange(context.Background(), code)
	if err != nil {
		fmt.Fprint(w, "failed")
		return
	}
	authServer.TokenChan <- token
	fmt.Fprint(w, "success!!")
}
