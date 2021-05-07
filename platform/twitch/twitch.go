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

func (t *twitch) GetTopStreamers(topX uint) ([]database.Streamer, error) {
	if topX > 100 {
		return nil, fmt.Errorf("topX must be <= 100")
	}

	url := fmt.Sprintf("%s?first=%d", streamsURL, topX)
	headers := http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", t.client.AccessToken)},
		"client-id":     []string{cfg.ClientID},
	}
	resp, err := t.client.Get(url, headers)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var data struct {
		Streams []*Stream `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(data.Streams))

	for _, stream := range data.Streams {
		userIDs = append(userIDs, stream.UserID)
	}

	streamerLookup, err := t.getChannelInfo(userIDs)
	if err != nil {
		return nil, err
	}

	if len(data.Streams) != len(streamerLookup) {
		fmt.Printf("MISSING %d CHANNEL INFO(S)\n", len(data.Streams)-len(streamerLookup))
	}

	streamers := make([]database.Streamer, 0, len(streamerLookup))
	for _, stream := range data.Streams {
		info, ok := streamerLookup[stream.UserID]
		if !ok {
			fmt.Printf("FAILED TO FETCH %s CHANNEL INFO\n", stream.UserLogin)
			continue
		}

		streamers = append(streamers, database.Streamer{
			ID:          stream.UserLogin,
			Username:    stream.UserName,
			Description: info.Description,
			Avatar:      info.ProfileImageURL,
			Platform:    database.Twitch,
		})
	}

	return streamers, nil
}

func (t *twitch) getChannelInfo(userIDs []string) (map[string]User, error) {
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
		Users []User `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	users := make(map[string]User)

	for _, user := range data.Users {
		users[user.ID] = user
	}

	return users, nil
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
