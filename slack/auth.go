package slack

import (
	"context"
	"fmt"
	"github.com/f110/montegrappa/db"
	"golang.org/x/oauth2"
	slackOAuth "golang.org/x/oauth2/slack"
	"log"
	"net/http"
)

type AuthServer struct {
	TeamID string
	conf   *oauth2.Config
}

func NewAuthServer(clientId, secret string, scopes []string, teamId string) *AuthServer {
	authServer := &AuthServer{
		TeamID: teamId,
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

func (authServer *AuthServer) Start(addr string) {
	log.Fatal(http.ListenAndServe(addr, nil))
}

func (authServer *AuthServer) startApp(w http.ResponseWriter, r *http.Request) {
	if t, _ := db.GetToken(); t != "" {
		fmt.Fprint(w, "already setup\n")
		fmt.Fprint(w, t)
		return
	}

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
	db.WriteToken(token)
	fmt.Fprint(w, "success!!")
}
