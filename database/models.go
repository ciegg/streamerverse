package database

import (
	"github.com/go-pg/pg/v10/orm"
	"time"
)

const (
	Twitch  Platform = "Twitch"
	YouTube          = "YouTube"
)

type Platform string

type Stream struct {
	Streamer
	Viewers []int64
}

type Streamer struct {
	ID                string `pg:",pk"`
	Username          string
	Platform          Platform
	Description       string
	CustomDescription string
	Avatar            string
	CustomAvatar      string
}

type viewer struct {
	// ORDER MATTERS
	Date       time.Time `pg:"type:DATE,notnull,unique:date_streamer_viewer"`
	StreamerID string    `pg:"unique:date_streamer_viewer"`
	ID         int64     `pg:"type:bigint,nopk,unique:date_streamer_viewer"`
}

func createSchema() error {
	models := []interface{}{
		(*Streamer)(nil),
		(*viewer)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	if _, err := db.Exec("SELECT create_hypertable('viewers', 'date', if_not_exists => TRUE)"); err != nil {
		return err
	}

	if _, err := db.Exec("CREATE INDEX ON viewers(streamer_id, date)"); err != nil {
		return err
	}

	return nil
}
