package dfeed

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
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

func ParseItem(siteTitle string, item *gofeed.Item, frequency int8) (*DFeed, error) {
	// interval to send feed
	interval := 24 / frequency

	if item.PublishedParsed == nil {
		return nil, errors.New("cannot get published date: " + item.Title)
	} else if item.PublishedParsed.Before(time.Now().Add(time.Duration(-(interval)) * time.Hour)) {
		return nil, errors.New("too old post: " + item.Title)
	} else if item.PublishedParsed.After(time.Now().Add(time.Duration(interval) * time.Hour)) {
		return nil, errors.New("too new post: " + item.Title)
	}

	return &DFeed{
		Title:     siteTitle,
		ItemTitle: item.Title,
		Url:       item.Link,
		Published: item.PublishedParsed,
	}, nil
}

func GetFeedConcurrently(wg *sync.WaitGroup, config Config, ch chan DFeed) {
	// First, add tasks.
	wg.Add(len(config.Feeds))
	fp := gofeed.NewParser()

	// wait for all goroutines to finish and close the channel.
	go func() {
		wg.Wait()
		fmt.Println("Completed: Closing channel.")
		close(ch)
	}()

	for _, feed := range config.Feeds {
		go func(feed string) {
			defer wg.Done()
			success := 0
			skipped := 0
			parsed, err := fp.ParseURL(feed)
			if err != nil {
				fmt.Println(err)
				return
			}
			items := parsed.Items
			for _, item := range items {
				d, err := ParseItem(parsed.Title, item, config.Frequency)
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
		}(feed)
	}
}

func SendFeed() {
	config := SetConfig()

	ch := make(chan DFeed)
	var wg sync.WaitGroup

	// Inside this function, tasks will be added, consumed and waited for their finishing.
	// Also data will be sent to channel, and the channel will be closed.
	GetFeedConcurrently(&wg, config, ch)

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

		req, err := http.NewRequest(http.MethodPost, config.Hook, bytes.NewReader(j))
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
