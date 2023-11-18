package dfeed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
)

type DFeed struct {
	Title     string     `json:"title"`
	ItemTitle string     `json:"item_title"`
	Url       string     `json:"url"`
	Published *time.Time `json:"published"`
}

type Req struct {
	Content string `json:"content"`
}

func ParseItem(siteTitle string, item *gofeed.Item) (*DFeed, error) {
	//Send feed 3 times in a day (24/3)
	timeLimit := 8

	if item.PublishedParsed == nil {
		return nil, errors.New("cannot get published date: " + item.Title)
	} else if item.PublishedParsed.Before(time.Now().Add(time.Duration(-(timeLimit)) * time.Hour)) {
		return nil, errors.New("too old post: " + item.Title)
	} else if item.PublishedParsed.After(time.Now().Add(time.Duration(timeLimit) * time.Hour)) {
		return nil, errors.New("too new post: " + item.Title)
	}

	return &DFeed{
		Title:     siteTitle,
		ItemTitle: item.Title,
		Url:       item.Link,
		Published: item.PublishedParsed,
	}, nil
}

func GetFeedConcurrently(feeds []string, ch chan DFeed) {
	defer close(ch)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	fp := gofeed.NewParser()

	for _, feed := range feeds {
		success := 0
		skipped := 0
		parsed, err := fp.ParseURLWithContext(feed, ctx)
		if err != nil {
			fmt.Println(err)
			continue
		}
		items := parsed.Items
		for _, item := range items {
			d, err := ParseItem(parsed.Title, item)
			if err != nil {
				skipped += 1
				continue
			}
			if d != nil {
				ch <- *d
				success += 1
			}
		}
		fmt.Println(parsed.Title + " SUCCESS: " + fmt.Sprint(success) + " SKIPPED: " + fmt.Sprint(skipped))
	}
}

func SendFeed(w http.ResponseWriter, r *http.Request) {
	feeds := SetFeedList()

	ch := make(chan DFeed)

	go GetFeedConcurrently(feeds, ch)

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
