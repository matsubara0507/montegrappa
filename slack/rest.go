package slack

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/f110/montegrappa/bot"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrChannelNotFound = errors.New("channel not found")
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

func (slackConnector *SlackConnector) GetUserInfo(userId string) (*UserInfo, error) {
	v := url.Values{}
	v.Set("token", slackConnector.token)
	v.Set("user", userId)

	res, err := http.PostForm("https://slack.com/api/users.info", v)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	resObj := new(UserInfoResponse)
	d.Decode(resObj)

	if resObj.Ok == false {
		return nil, ErrUserNotFound
	}

	return &resObj.User, nil
}

func (slackConnector *SlackConnector) GetChannelInfo(channelId string) (*bot.ChannelInfo, error) {
	v := url.Values{}
	v.Set("token", slackConnector.token)
	v.Set("channel", channelId)
	res, err := http.PostForm("https://slack.com/api/channels.info", v)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	resObj := new(ChannelInfoResponse)
	d.Decode(resObj)

	if resObj.Ok == false {
		return nil, ErrChannelNotFound
	}

	return &resObj.Channel, nil
}

func (slackConnector *SlackConnector) GetTeamInfo() (*TeamInfo, error) {
	v := url.Values{}
	v.Set("token", slackConnector.token)
	res, err := http.PostForm("https://slack.com/api/team.info", v)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := json.NewDecoder(res.Body)
	resObj := new(TeamInfoResponse)
	d.Decode(resObj)

	if resObj.Ok == false {
		return nil, errors.New("something wrong")
	}

	return &resObj.Team, nil
}
