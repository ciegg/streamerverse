package platform

import "streamerverse/database"

type Platform interface {
	GetTopStreamers(uint) ([]database.Streamer, error)
	GetViewers(string) ([]int64, error)
	Name() database.Platform
}
