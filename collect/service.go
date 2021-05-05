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
	topX      uint
	jobs      chan *database.Stream
	platforms []platform.Platform
}

func NewCollector(platforms []platform.Platform) *collector {
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
			fmt.Printf("Fetching top %d streamers on %s\n", c.topX, plt.Name())

			var streamers []database.Streamer
			err := withRetry(3, func() error {
				var err error
				streamers, err = plt.GetTopStreamers(c.topX)
				return err
			})

			if err != nil {
				fmt.Printf("FAILED TO FETCH TOP %d STREAMERS FOR %s: %s", c.topX, plt.Name(), err)
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
				go c.worker(&wg, x, plt)
			}

			wg.Wait()

			fmt.Printf("Finished inserting top %d streams for %s", c.topX, plt.Name())
		}

		time.Sleep(c.interval)
	}
}

func (c *collector) worker(wg *sync.WaitGroup, num int, plt platform.Platform) {
	defer wg.Done()

	for {
		if len(c.jobs) == 0 {
			break
		}

		stream := <-c.jobs

		fmt.Printf("WORKER %d: Fetching %s viewers on %s\n", num, stream.Name, plt.Name())

		err := withRetry(3, func() error {
			var err error
			stream.Viewers, err = plt.GetViewers(stream.Name)
			return err
		})
		if err != nil {
			fmt.Printf("FAILED TO %s VIEWERS: %s\n", stream.Name, err)
			continue
		}

		fmt.Printf("WORKER %d: Inserting %s viewers into db\n", num, stream.Name)

		err = withRetry(3, func() error {
			return database.Insert(stream)
		})
		if err != nil {
			fmt.Printf("FAILED TO INSERT %s INTO DB ON %s: %s\n", stream.Name, time.Now().String(), err)
		} else {
			fmt.Printf("WORKER %d: Finished inserting %s viewers\n", num, stream.Name)
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
