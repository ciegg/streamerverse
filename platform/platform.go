package platform

import "streamerverse/database"

type Platform interface {
	GetTopStreamers(int) ([]database.Streamer, error)
	GetViewers(string) ([]int64, error)
	Name() database.Platform
}
