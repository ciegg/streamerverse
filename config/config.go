package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type config struct {
	DatabaseURI string
	ClientID    string
	Secret      string
	Debug       bool
	Interval    Duration
	TopX        int
}

var Config = &config{}

func LoadConfig() error {
	fmt.Println("Loading config")

	configFile, err := os.Open("config.json")
	if err != nil {
		return err
	}

	defer configFile.Close()

	if err := json.NewDecoder(configFile).Decode(Config); err != nil {
		return err
	}

	if Config.TopX == 0 {
		Config.TopX = 100
	}

	if Config.Interval.Duration == 0 {
		Config.Interval.Duration = time.Hour
	}

	fmt.Println("Config loaded")

	return nil
}
