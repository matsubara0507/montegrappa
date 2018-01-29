package slack

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/f110/montegrappa/bot"
)

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrChannelNotFound          = errors.New("channel not found")
	ErrFailedPostMessage        = errors.New("Failed post message")
	ErrFailedGetRTMEndpoint     = errors.New("Failed getting RTM Endpoind")
	ErrFailedOpenPrivateChannel = errors.New("Failed open private channel")
)

type UserInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type TeamInfo struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type User struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
}

type Channel struct {
	Id string `json:"id"`
}

type UserInfoResponse struct {
	Ok   bool     `json:"ok"`
	User UserInfo `json:"user"`
}

type ChannelInfoResponse struct {
	Ok      bool            `json:"ok"`
	Channel bot.ChannelInfo `json:"channel"`
}

type TeamInfoResponse struct {
	Ok   bool     `json:"ok"`
	Team TeamInfo `json:"team"`
}

type UserListResponse struct {
	Ok      bool   `json:"ok"`
	Members []User `json:"members"`
}

type PostMesssageResponse struct {
	Ok bool   `json:"ok"`
	Ts string `json:"ts"`
}

type RTMConnectResponse struct {
	Ok  bool   `json:"ok"`
	URL string `json:"url"`
}

type IMOpenResponse struct {
	Ok      bool    `json:"ok"`
	Channel Channel `json:"channel"`
}

func (slackConnector *SlackConnector) PostMessage(channel, text, username string) (*PostMesssageResponse, error) {
	v := url.Values{}
	v.Set("channel", channel)
	v.Set("text", text)
	v.Set("as_user", "false")
	if username != "" {
		v.Set("username", username)
	}

	res, err := slackConnector.callRestAPI(context.Background(), "chat.postMessage", v)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	var resObj PostMesssageResponse
	err = d.Decode(&resObj)
	if err != nil {
		return nil, err
	}

	if resObj.Ok == false {
		return nil, ErrFailedPostMessage
	}

	return &resObj, nil
}

func (slackConnector *SlackConnector) GetUserInfo(userId string) (*UserInfo, error) {
	v := url.Values{}
	v.Set("user", userId)
	res, err := slackConnector.callRestAPI(context.Background(), "users.info", v)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	var resObj UserInfoResponse
	err = d.Decode(&resObj)
	if err != nil {
		return nil, err
	}

	if resObj.Ok == false {
		return nil, ErrUserNotFound
	}

	return &resObj.User, nil
}

func (slackConnector *SlackConnector) GetChannelInfo(channelId string) (*bot.ChannelInfo, error) {
	v := url.Values{}
	v.Set("channel", channelId)
	res, err := slackConnector.callRestAPI(context.Background(), "channels.info", v)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	var resObj ChannelInfoResponse
	err = d.Decode(&resObj)
	if err != nil {
		return nil, err
	}

	if resObj.Ok == false {
		return nil, ErrChannelNotFound
	}

	return &resObj.Channel, nil
}

func (slackConnector *SlackConnector) GetTeamInfo() (*TeamInfo, error) {
	res, err := slackConnector.callRestAPI(context.Background(), "team.info", url.Values{})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	var resObj TeamInfoResponse
	err = d.Decode(&resObj)
	if err != nil {
		return nil, err
	}

	if resObj.Ok == false {
		return nil, errors.New("something wrong")
	}

	return &resObj.Team, nil
}

func (slackConnector *SlackConnector) GetUserList() ([]User, error) {
	v := url.Values{}
	v.Set("presence", "false")
	res, err := slackConnector.callRestAPI(context.Background(), "users.list", v)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	var resObj UserListResponse
	err = d.Decode(&resObj)
	if err != nil {
		return nil, err
	}

	if resObj.Ok == false {
		return nil, errors.New("can not get users.list")
	}

	return resObj.Members, nil
}

func (slackConnector *SlackConnector) RTMConnect() (string, error) {
	res, err := slackConnector.callRestAPI(context.Background(), "rtm.connect", url.Values{})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	var resObj RTMConnectResponse
	err = d.Decode(&resObj)
	if err != nil {
		return "", err
	}

	if resObj.Ok == false {
		return "", ErrFailedGetRTMEndpoint
	}

	if resObj.URL == "" {
		return "", ErrFailedGetRTMEndpoint
	}

	return resObj.URL, nil
}

func (slackConnector *SlackConnector) IMOpen(userId string) (*Channel, error) {
	v := url.Values{}
	v.Set("user", userId)
	res, err := slackConnector.callRestAPI(context.Background(), "im.open", v)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	var resObj IMOpenResponse
	err = d.Decode(&resObj)
	if err != nil {
		return nil, err
	}

	if resObj.Ok == false {
		return nil, ErrFailedOpenPrivateChannel
	}

	return &resObj.Channel, nil
}

func (slackConnector *SlackConnector) callRestAPI(ctx context.Context, method string, v url.Values) (*http.Response, error) {
	v.Set("token", slackConnector.token)
	b := strings.NewReader(v.Encode())

	req, err := http.NewRequest("POST", "https://slack.com/api/"+method, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqWithContext := req.WithContext(ctx)
	return http.DefaultClient.Do(reqWithContext)
}
