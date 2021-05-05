package database

import (
	"fmt"
	"github.com/go-pg/pg/v10/orm"
)

const (
	Twitch  Platform = "Twitch"
	YouTube          = "YouTube"
)

type Platform string

type Stream struct {
	Streamer
	Viewers UintSlice
}

type Streamer struct {
	Name              string `pg:",pk"`
	Platform          Platform
	Description       string
	CustomDescription string
	Avatar            string
	CustomAvatar      string
}

func createSchema() error {
	models := []interface{}{
		(*Streamer)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func createViewerTable() error {
	viewersTable := getCurrentViewersTable()

	fmt.Printf("Adding new viewers table: %s\n", viewersTable)

	tableSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id numeric(20),
	streamer text,
	PRIMARY KEY(id, streamer)
);
`, viewersTable)
	_, err := db.Exec(tableSQL)

	return err
}
