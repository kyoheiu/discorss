package feed_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyoheiu/discorss/feed"
	"github.com/mmcdole/gofeed"
)

func TestParseFeed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	feeds := feed.SetFeedList()

	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(feeds[0], ctx)
	if err != nil {
		t.Error("Cannot get or parse feed")
	}
	items := feed.Items
	for _, item := range items {
		if item.PublishedParsed == (*time.Time)(nil) {
			t.Log("Cannot get published date: ", item.Title)
			continue
		} else if item.PublishedParsed.Before(time.Now().Add(time.Duration(-24) * time.Hour)) {
			t.Log("Too old: ", item.Title)
			continue
		} else if item.PublishedParsed.After(time.Now().Add(time.Duration(24) * time.Hour)) {
			t.Log("Too new: ", item.Title)
			continue
		}

		t.Error(item.Title)
	}
}
