package slack

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type UserInfoResponse struct {
	Ok   bool     `json:"ok"`
	User UserInfo `json:"user"`
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
