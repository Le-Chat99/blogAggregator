package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Le-Chat99/blogAggregator/internal/database"
	"github.com/google/uuid"
)

const date_format = "Mon, 02 Jan 2006 15:04:05 -0700"

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FeedFetcher(feedUrl string) (*RSSFeed, error) {
	Client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := Client.Get(feedUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data RSSFeed
	err = xml.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil

}

func FectchOne(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()
	_, err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed %s fetched: %v", feed.Name, err)
		return
	}

	feedData, err := FeedFetcher(feed.Url)
	if err != nil {
		log.Printf("Couldn't collect feed %s: %v", feed.Name, err)
		return
	}

	for _, item := range feedData.Channel.Item {
		des := sql.NullString{
			String: item.Description,
			Valid:  true,
		}
		pubDate, err := time.Parse(date_format, item.PubDate)
		if err != nil {
			log.Printf("Couldn't parse pubDate %s: %v", item.PubDate, err)
			continue
		}
		date := sql.NullTime{
			Time:  pubDate,
			Valid: true,
		}
		postCreat := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: des,
			PublishedAt: date,
			FeedID:      feed.ID,
		}
		_, err = db.CreatePost(context.Background(), postCreat)
		if err != nil {
			log.Printf("Couldn't create post: %v", err)
			continue
		}
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}

func FecthNTime(db *database.Queries, concurrency int, requestDelayTime time.Duration) {
	log.Printf("Collecting feeds every %v on %v goroutines...", requestDelayTime, concurrency)
	ticker := time.NewTicker(requestDelayTime)
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.Println("Couldn't get next feeds to fetch", err)
			continue
		}
		log.Printf("Found %v feeds to fetch!", len(feeds))

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)
			go FectchOne(db, wg, feed)
		}
		wg.Wait()
	}
}
