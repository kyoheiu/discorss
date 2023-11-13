package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

type DFeed struct {
	Title     string
	ItemTitle string
	ItemDesc  string
	Url       string
	Published *time.Time
}

func parseFeed(wg *sync.WaitGroup, feeds []string, ch chan DFeed) {
	defer wg.Done()
	defer close(ch)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fp := gofeed.NewParser()
	for _, v := range feeds {
		if len(v) == 0 {
			continue
		}
		feed, err := fp.ParseURLWithContext(v, ctx)
		if err != nil {
			fmt.Println("Cannot get or parse feed: ", v)
		}
		items := feed.Items
		for _, item := range items {
			var desc string

			if item.PublishedParsed == (*time.Time)(nil) {
				continue
			} else if item.PublishedParsed.Before(time.Now().Add(time.Duration(-24) * time.Hour)) {
				continue
			}

			if len(item.Description) >= 50 {
				desc = item.Description[:50]
			} else {
				desc = item.Description
			}
			ch <- DFeed{
				Title:     feed.Title,
				ItemTitle: item.Title,
				ItemDesc:  desc,
				Url:       item.Link,
				Published: item.PublishedParsed,
			}
		}
	}
}

func main() {
	f, err := os.ReadFile("./python/feed.txt")
	if err != nil {
		panic("Cannot read text file.")
	}
	str := string(f)
	feeds := strings.Split(str, "\n")

	ch := make(chan DFeed)

	var wg sync.WaitGroup
	wg.Add(1)

	go parseFeed(&wg, feeds, ch)

	for dfeed := range ch {
		fmt.Println(dfeed)
	}
}
