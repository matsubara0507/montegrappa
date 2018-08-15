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
	ErrFailedPostMessage        = errors.New("failed post message")
	ErrFailedGetRTMEndpoint     = errors.New("failed getting RTM Endpoint")
	ErrFailedOpenPrivateChannel = errors.New("failed open private channel")
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
	Id         string `json:"id"`
	Name       string `json:"name"`
	IsChannel  bool   `json:"is_channel"`
	IsGroup    bool   `json:"is_group"`
	IsIM       bool   `json:"is_im"`
	IsMember   bool   `json:"is_member"`
	IsPrivate  bool   `json:"is_private"`
	IsGeneral  bool   `json:"is_general"`
	IsMPIM     bool   `json:"is_mpim"`
	NumMembers int    `json:"num_members"`
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

type PostMessageResponse struct {
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

type ConversationListResponse struct {
	Ok       bool      `json:"ok"`
	Channels []Channel `json:"channels"`
}

func (connector *Connector) PostMessage(channel, text, username string) (*PostMessageResponse, error) {
	v := url.Values{}
	v.Set("channel", channel)
	v.Set("text", text)
	v.Set("as_user", "false")
	if username != "" {
		v.Set("username", username)
	}

	res, err := connector.callRestAPI(context.Background(), "chat.postMessage", v)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	var resObj PostMessageResponse
	err = d.Decode(&resObj)
	if err != nil {
		return nil, err
	}

	if resObj.Ok == false {
		return nil, ErrFailedPostMessage
	}

	return &resObj, nil
}

func (connector *Connector) GetUserInfo(userId string) (*UserInfo, error) {
	v := url.Values{}
	v.Set("user", userId)
	res, err := connector.callRestAPI(context.Background(), "users.info", v)
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

func (connector *Connector) GetChannelInfo(channelId string) (*bot.ChannelInfo, error) {
	v := url.Values{}
	v.Set("channel", channelId)
	res, err := connector.callRestAPI(context.Background(), "channels.info", v)
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

func (connector *Connector) GetTeamInfo() (*TeamInfo, error) {
	res, err := connector.callRestAPI(context.Background(), "team.info", url.Values{})
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

func (connector *Connector) GetUserList() ([]User, error) {
	v := url.Values{}
	v.Set("presence", "false")
	res, err := connector.callRestAPI(context.Background(), "users.list", v)
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

func (connector *Connector) RTMConnect() (string, error) {
	res, err := connector.callRestAPI(context.Background(), "rtm.connect", url.Values{})
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

func (connector *Connector) IMOpen(userId string) (*Channel, error) {
	v := url.Values{}
	v.Set("user", userId)
	res, err := connector.callRestAPI(context.Background(), "im.open", v)
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

func (connector *Connector) GetJoinedChannelList() ([]Channel, error) {
	v := url.Values{}
	v.Set("exclude_archived", "true")
	v.Set("types", "public_channel,private_channel")
	res, err := connector.callRestAPI(context.Background(), "conversations.list", v)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var resObj ConversationListResponse
	if err := json.NewDecoder(res.Body).Decode(&resObj); err != nil {
		return nil, err
	}

	channels := make([]Channel, 0)
	for _, c := range resObj.Channels {
		if c.IsMember {
			channels = append(channels, c)
		}
	}

	return channels, nil
}

func (connector *Connector) callRestAPI(ctx context.Context, method string, v url.Values) (*http.Response, error) {
	v.Set("token", connector.token)
	b := strings.NewReader(v.Encode())

	req, err := http.NewRequest("POST", "https://slack.com/api/"+method, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqWithContext := req.WithContext(ctx)
	return http.DefaultClient.Do(reqWithContext)
}
