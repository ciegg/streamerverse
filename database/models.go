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
	tableName struct{} `pg:"viewers,partition_by:LIST(iso_week)"`

	// ORDER MATTERS
	ISOWeek    string `pg:",type:varchar(9),pk"`
	StreamerID string `pg:",pk"`
	ID         int64  `pg:"type:bigint,pk"`
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

	return nil
}

func createNewViewerPartition() error {
	viewersTable, year, week := getCurrentViewersTable()

	fmt.Printf("Adding new viewers table: %s\n", viewersTable)

	tableSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s PARTITION OF viewers FOR VALUES IN ('%s');`, viewersTable, fmt.Sprintf("Y%d-W%d", year, week))
	_, err := db.Exec(tableSQL)

	return err
}
