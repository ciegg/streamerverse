package collect

import (
	"fmt"
	"streamerverse/config"
	"streamerverse/database"
	"streamerverse/platform"
	"sync"
	"time"
)

var cfg = config.Config

type collector struct {
	interval  time.Duration
	topX      int
	jobs      chan *database.Stream
	platforms []platform.Platform
}

func NewCollector(platforms ...platform.Platform) *collector {
	return &collector{
		interval:  cfg.Interval.Duration,
		topX:      cfg.TopX,
		jobs:      make(chan *database.Stream, cfg.TopX),
		platforms: platforms,
	}
}

func (c *collector) Start() {
	fmt.Printf("Starting collector with %d platforms on %s interval\n", len(c.platforms), c.interval.String())
	for {
		for _, plt := range c.platforms {
			start := time.Now()

			year, week := start.ISOWeek()
			bucket := isoWeekToTime(year, week)

			fmt.Printf("Fetching top %d streamers on %s\n", c.topX, plt.Name())

			var streamers []database.Streamer
			err := withRetry(3, func() error {
				var err error
				streamers, err = plt.GetTopStreamers(c.topX)
				return err
			})

			if err != nil {
				fmt.Printf("FAILED TO FETCH TOP %d STREAMERS FOR %s: %s\n", c.topX, plt.Name(), err)
				continue
			}

			streams := make([]*database.Stream, 0, len(streamers))

			for _, streamer := range streamers {
				stream := &database.Stream{
					Streamer: streamer,
				}

				streams = append(streams, stream)

				c.jobs <- stream
			}

			var wg sync.WaitGroup

			for x := 0; x < 10; x++ {
				wg.Add(1)
				go c.worker(&wg, x, plt, bucket)
			}

			wg.Wait()

			fmt.Printf("Finished inserting top %d streams for %s: %s\n", c.topX, plt.Name(), time.Now().Sub(start))
		}

		fmt.Printf("Next fetch at %s\n", time.Now().Add(c.interval))
		time.Sleep(c.interval)
	}
}

func (c *collector) worker(wg *sync.WaitGroup, num int, plt platform.Platform, bucket time.Time) {
	defer wg.Done()

	for {
		if len(c.jobs) == 0 {
			break
		}

		stream := <-c.jobs

		fmt.Printf("WORKER %d: Fetching %s viewers on %s\n", num, stream.Username, plt.Name())

		err := withRetry(3, func() error {
			var err error
			stream.Viewers, err = plt.GetViewers(stream.ID)
			return err
		})
		if err != nil {
			fmt.Printf("FAILED TO %s VIEWERS: %s\n", stream.Username, err)
			continue
		}

		fmt.Printf("WORKER %d: Inserting %s viewers into db\n", num, stream.Username)

		err = withRetry(3, func() error {
			return database.Insert(stream, bucket)
		})
		if err != nil {
			fmt.Printf("INSERTING %s INTO DB FAILED: %s\n", stream.Username, err)
		} else {
			fmt.Printf("WORKER %d: Finished inserting %s viewers\n", num, stream.Username)
		}
	}
}

func withRetry(num int, op func() error) error {
	var err error
	for x := num; x > 0; x-- {
		err = op()
		if err == nil {
			break
		}
	}

	return err
}

// Taken from https://stackoverflow.com/a/52303730
func isoWeekToTime(year, week int) time.Time {
	// Start from the middle of the year:
	t := time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC)

	// Roll back to Monday:
	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	// Difference in weeks:
	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (week-w)*7)

	return t
}
