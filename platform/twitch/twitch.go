package twitch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"streamerverse/config"
	"streamerverse/database"
	"streamerverse/platform"
	"strings"
)

const (
	chatroomURL = "https://tmi.twitch.tv/group/user"

	apiURL     = "https://api.twitch.tv/helix"
	streamsURL = apiURL + "/streams"
	userURL    = apiURL + "/users"
)

var cfg = config.Config

type twitch struct {
	client twitchClient
}

func Setup() (*twitch, error) {
	fmt.Println("Setting up twitch interface")

	t := &twitch{
		client: twitchClient{
			Client: platform.NewClient(),
		},
	}

	auth, err := t.client.getAppToken()
	if err != nil {
		return nil, err
	}

	t.client.auth = *auth

	return t, nil
}

func (t *twitch) GetTopStreamers(topX int) ([]database.Streamer, error) {
	var userIDs []string
	for x := topX; x < len(channels); x += topX {
		url := fmt.Sprintf("%s?user_login=%s", streamsURL, strings.Join(channels[x-100:x], "&user_login="))
		headers := http.Header{
			"Authorization": []string{fmt.Sprintf("Bearer %s", t.client.AccessToken)},
			"client-id":     []string{cfg.ClientID},
		}
		resp, err := t.client.Get(url, headers)
		if err != nil {
			return nil, err
		}

		var data struct {
			Streams []*Stream `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, err
		}

		resp.Body.Close()

		for _, stream := range data.Streams {
			userIDs = append(userIDs, stream.UserID)
			if len(userIDs) >= topX {
				break
			}
		}

		if len(userIDs) >= topX {
			break
		}
	}

	streamerLookup, err := t.getChannelInfo(userIDs)
	if err != nil {
		return nil, err
	}

	var streamers []database.Streamer
	for _, user := range streamerLookup {
		streamers = append(streamers, database.Streamer{
			ID:          user.Login,
			Username:    user.DisplayName,
			Description: user.Description,
			Avatar:      user.ProfileImageURL,
			Platform:    database.Twitch,
		})
	}

	return streamers, nil
}

func (t *twitch) getChannelInfo(userIDs []string) ([]*User, error) {
	url := fmt.Sprintf("%s?id=%s", userURL, strings.Join(userIDs, "&id="))
	headers := http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", t.client.AccessToken)},
		"Client-Id":     []string{cfg.ClientID},
	}

	resp, err := t.client.Get(url, headers)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var data struct {
		Users []*User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Users, nil
}

func (t *twitch) GetViewers(username string) ([]int64, error) {
	resp, err := t.client.Client.Get(fmt.Sprintf("%s/%s/chatters", chatroomURL, username), nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var chatroom struct {
		ChatterCount uint `json:"chatter_count"`
		Chatters     struct {
			Broadcaster []string `json:"broadcaster"`
			VIPs        []string `json:"vips"`
			Moderators  []string `json:"moderators"`
			Staff       []string `json:"staff"`
			Admins      []string `json:"admins"`
			GlobalMods  []string `json:"global_mods"`
			Viewers     []string `json:"viewers"`
		} `json:"chatters"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&chatroom); err != nil {
		return nil, err
	}

	chatters := make([]string, 0, chatroom.ChatterCount-1)

	chatters = append(chatters, chatroom.Chatters.VIPs...)
	chatters = append(chatters, chatroom.Chatters.Moderators...)
	chatters = append(chatters, chatroom.Chatters.Staff...)
	chatters = append(chatters, chatroom.Chatters.Admins...)
	chatters = append(chatters, chatroom.Chatters.GlobalMods...)
	chatters = append(chatters, chatroom.Chatters.Viewers...)

	return database.HashViewers(chatters), nil
}

func (t *twitch) Name() database.Platform {
	return database.Twitch
}
