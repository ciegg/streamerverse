package platform

import "streamerverse/database"

type Platform interface {
	GetTopStreamers(uint) ([]database.Streamer, error)
	GetViewers(string) (database.UintSlice, error)
	Name() database.Platform
}
