package feed_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	dfeed "github.com/kyoheiu/discorss/dfeed"
	"github.com/mmcdole/gofeed"
)

func TestParseFeed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	config := dfeed.SetConfig()

	fp := gofeed.NewParser()
	for _, f := range config.Feeds {
		feed, err := fp.ParseURLWithContext(f, ctx)
		if err != nil {
			t.Log("cannot get or parse feed: " + f)
			return
		}
		items := feed.Items
		for _, item := range items {
			d, err := dfeed.ParseItem(feed.Title, item, config.Frequency)
			if err != nil {
				t.Log(err)
				continue
			}
			t.Log("Success: " + d.ItemTitle)
		}
	}
}

func TestEmptyFeed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fp := gofeed.NewParser()
	_, err := fp.ParseURLWithContext("", ctx)
	if err != nil {
		t.Log("cannot get or parse feed of empty URL")
		return
	}
}

func TestGetFeedConcurrently(t *testing.T) {
	config := dfeed.SetConfig()
	ch := make(chan dfeed.DFeed)
	var wg sync.WaitGroup
	dfeed.GetFeedConcurrently(&wg, config, ch)
	for f := range ch {
		fmt.Println(f.ItemTitle)
	}
}
