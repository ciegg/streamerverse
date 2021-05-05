package twitch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"streamerverse/platform"
	"time"
)

const oauthURL = "https://id.twitch.tv/oauth2/token"

type auth struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   uint   `json:"expires_in"`

	// internally  specified
	ExpiresAt time.Time
}

type twitchClient struct {
	*platform.Client
	auth
}

func (t *twitchClient) getAppToken() (*auth, error) {
	url := fmt.Sprintf("%s?client_id=%s&client_secret=%s&grant_type=client_credentials", oauthURL, cfg.ClientID, cfg.Secret)
	resp, err := t.Client.Post(url, nil, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data := &auth{}
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}

	// 1000000000 nanoseconds in a second
	data.ExpiresAt = time.Now().Add(time.Duration(data.ExpiresIn * 1000000000))

	return data, nil
}

func (t *twitchClient) Get(url string, headers http.Header) (*http.Response, error) {
	if time.Now().After(t.ExpiresAt) {
		auth, err := t.getAppToken()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh app token: %s", err)
		}

		t.auth = *auth

		if _, ok := headers["Authorization"]; ok {
			headers.Set("Authorization", fmt.Sprintf("Bearer %s", t.AccessToken))
		}
	}

	return t.Client.Get(url, headers)
}
