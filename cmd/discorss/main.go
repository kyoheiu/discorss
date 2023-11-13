package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

type DFeed struct {
	Title     string     `json:"title"`
	ItemTitle string     `json:"item_title"`
	ItemDesc  string     `json:"item_desc"`
	Url       string     `json:"url"`
	Published *time.Time `json:"published"`
}

type Req struct {
	Content string `json:"content"`
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
			} else if item.PublishedParsed.After(time.Now().Add(time.Duration(24) * time.Hour)) {
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

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	for dfeed := range ch {
		content := fmt.Sprintf("%s %s <%s>", dfeed.Title, dfeed.ItemTitle, dfeed.Url)
		j, err := json.Marshal(Req{Content: content})
		if err != nil {
			fmt.Println(err)
			continue
		}

		url := os.Getenv("DISCORSS_URL")
		if len(url) == 0 {
			fmt.Println("Cannot get webhook url.")
			continue
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(j))
		if err != nil {
			fmt.Println(err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(dfeed.ItemTitle, dfeed.Url, resp.StatusCode)
	}
}
