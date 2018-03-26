package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	slackOAuth "golang.org/x/oauth2/slack"
)

type AuthServer struct {
	TeamID    string
	TokenChan chan *Token
	conf      *oauth2.Config
	server    *http.Server
	mutex     sync.Mutex
}

type Token struct {
	AccessToken    string    `json:"access_token"`
	RefreshToken   string    `json:"refresh_token,omitempty"`
	Expiry         time.Time `json:"expiry,omitempty"`
	BotUserID      string    `json:"bot_user_id,omitempty"`
	BotAccessToken string    `json:"bot_access_token,omitempty"`
}

func NewAuthServer(clientId, secret string, scopes []string, redirectUrl string, teamId string) *AuthServer {
	authServer := &AuthServer{
		TeamID:    teamId,
		TokenChan: make(chan *Token, 1),
		conf: &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: secret,
			Scopes:       scopes,
			Endpoint:     slackOAuth.Endpoint,
			RedirectURL:  redirectUrl,
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

func (authServer *AuthServer) Shutdown(ctx context.Context) error {
	return authServer.server.Shutdown(ctx)
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
	slackToken := Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
	bot, ok := token.Extra("bot").(map[string]interface{})
	if ok {
		slackToken.BotAccessToken = bot["bot_access_token"].(string)
		slackToken.BotUserID = bot["bot_user_id"].(string)
	}
	authServer.TokenChan <- &slackToken
	fmt.Fprint(w, "success!!")
}
