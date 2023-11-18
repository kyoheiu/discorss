package dfeed

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
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

	var desc string
	if len(item.Description) >= 50 {
		desc = item.Description[:50]
	} else {
		desc = item.Description
	}
	return &DFeed{
		Title:     siteTitle,
		ItemTitle: item.Title,
		ItemDesc:  desc,
		Url:       item.Link,
		Published: item.PublishedParsed,
	}, nil
}

func AddFeedToChannel(feeds []string, ch chan DFeed) {
	var wg sync.WaitGroup
	defer close(ch)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fp := gofeed.NewParser()

	for _, feed := range feeds {
		f := feed
		wg.Add(1)
		go func() {
			defer wg.Done()
			parsed, err := fp.ParseURLWithContext(f, ctx)
			if err != nil {
				fmt.Println(err)
				return
			}
			items := parsed.Items
			for _, item := range items {
				d, err := ParseItem(parsed.Title, item)
				if err != nil {
					fmt.Println(err)
					continue
				}
				ch <- *d
			}
		}()
	}

	wg.Wait()
}

func SendFeed(w http.ResponseWriter, r *http.Request) {
	feeds := SetFeedList()

	ch := make(chan DFeed)
	defer close(ch)

	AddFeedToChannel(feeds, ch)

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
