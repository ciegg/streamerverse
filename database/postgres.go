package database

import (
	"context"
	"fmt"
	"github.com/go-pg/pg/extra/pgdebug"
	"streamerverse/config"
	"time"

	"github.com/go-pg/pg/v10"
)

var db *pg.DB

func ConnectToDB() error {
	opt, err := pg.ParseURL(config.Config.DatabaseURI)
	if err != nil {
		return err
	}

	fmt.Println("Connecting to database")

	db = pg.Connect(opt)

	err = db.Ping(context.Background())
	if err != nil {
		return err
	}

	if config.Config.Debug {
		db.AddQueryHook(&pgdebug.DebugHook{
			Verbose: true,
		})
	}

	if err := createSchema(); err != nil {
		return err
	}

	fmt.Println("Connected to database")

	return nil
}

func CloseDB() {
	if err := db.Close(); err != nil {
		fmt.Println(err)
	}
}

func Insert(stream *Stream, now time.Time) error {
	viewers := make([]viewer, 0, len(stream.Viewers))

	for _, chatter := range stream.Viewers {
		viewers = append(viewers, viewer{
			ID:         chatter,
			StreamerID: stream.ID,
			Date:       now,
		})
	}

	_, err := db.Model(&viewers).OnConflict("DO NOTHING").Insert()
	if err != nil {
		fmt.Printf("FAILED TO INSERT %s VIEWERS: %s\n", stream.Username, err)
	}

	_, err = db.Model(&stream.Streamer).
		OnConflict("(id) DO UPDATE").
		Set("username = EXCLUDED.username, description = EXCLUDED.description, avatar = EXCLUDED.avatar").
		Insert()

	if err != nil {
		err = fmt.Errorf("failed to upsert %s: %s", stream.Username, err)
	}

	return err
}
