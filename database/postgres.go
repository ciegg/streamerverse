package database

import (
	"context"
	"fmt"
	"github.com/go-pg/pg/extra/pgdebug"
	"streamerverse/config"
	"strings"
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

func Insert(stream *Stream) error {
	viewersTable, year, week := getCurrentViewersTable()
	_, err := db.Exec(fmt.Sprintf(`SELECT 1 FROM %s`, viewersTable))

	// Postgres undefined_table
	// https://www.postgresql.org/docs/11/errcodes-appendix.html
	if err != nil && strings.Contains(err.Error(), "42P01") {
		if err := createNewViewerPartition(); err != nil {
			return fmt.Errorf("failed to create new viewers table: %s", err)
		}
	} else if err != nil {
		return err
	}

	viewers := make([]viewer, 0, len(stream.Viewers))

	for _, chatter := range stream.Viewers {
		viewers = append(viewers, viewer{
			ID:         chatter,
			StreamerID: stream.ID,
			ISOWeek:    fmt.Sprintf("Y%d-W%d", year, week),
		})
	}

	_, err = db.Model(&viewers).OnConflict("DO NOTHING").Insert()
	if err != nil {
		fmt.Printf("FAILED TO INSERT %s VIEWERS INTO %s: %s\n", stream.Username, viewersTable, err)
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

func getCurrentViewersTable() (string, int, int) {
	year, week := time.Now().ISOWeek()

	return fmt.Sprintf("viewers_y%d_w%d", year, week), year, week
}
