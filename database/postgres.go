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
	viewersTable := getCurrentViewersTable()
	_, err := db.Exec(fmt.Sprintf(`SELECT 1 FROM %s`, viewersTable))

	// Postgres undefined_table
	// https://www.postgresql.org/docs/11/errcodes-appendix.html
	if err != nil && strings.Contains(err.Error(), "42P01") {
		if err := createViewerTable(); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	for _, chatter := range stream.Viewers {
		viewer := map[string]interface{}{
			"id":       chatter,
			"streamer": stream.Name,
		}

		_, err := db.Model(&viewer).TableExpr(viewersTable).OnConflict("DO NOTHING").Insert()
		if err != nil {
			fmt.Printf("FAILED TO INSERT %s VIEWER %d INTO %s: %s\n", stream.Name, chatter, viewersTable, err)
		}
	}

	_, err = db.Model(&stream.Streamer).
		OnConflict("(name) DO UPDATE").
		Set("description = EXCLUDED.description, avatar = EXCLUDED.avatar").
		Insert()
	if err != nil {
		return err
	}

	return nil
}

func getCurrentViewersTable() string {
	year, week := time.Now().ISOWeek()

	return fmt.Sprintf("viewers_%d_%d", year, week)
}
