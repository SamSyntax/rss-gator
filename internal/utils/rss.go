package utils

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/SamSyntax/Gator/internal/database"
	"github.com/google/uuid"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func StartScrape(db *database.Queries, concurrency int, timeBetweenScrape time.Duration) {
	log.Printf("Scraping on %v goroutines every %s duration\n", concurrency, timeBetweenScrape)
	ticker := time.NewTicker(timeBetweenScrape)

	for ; ; <-ticker.C {
    log.Print("Starting another batch\n")
		feeds, err := db.GetNextFeedsToFetch(context.TODO(), int32(concurrency))
		if err != nil {
			log.Printf("Error getting next feeds to fetch %v\n", err)
			continue
		}

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)
			go scrapeFeeds(db, wg, feed)
		}
		wg.Wait()
	}
}

func scrapeFeeds(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()
	ctx := context.Background()

	_, err := db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		log.Fatalf("Couldn't mark feed as fetched: %v", err)
	}
	rssFeed, err := FetchFeed(ctx, feed.Url)
	if err != nil {
		log.Fatalf("Couldn't scrape feed: %v", err)
	}
	for _, item := range rssFeed.Channel.Item {
		desc := sql.NullString{}
		if item.Description != "" {
			desc.String = item.Description
			desc.Valid = true
		}
		t, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("Error parsing publish date: %v\n", err)
		}
		if item.Title == "" {
			log.Printf("Title of %v is empty, skipping\n", item.Link)
		}
		_, err = db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			UpdatedAt:   time.Now(),
			CreatedAt:   time.Now(),
			FeedID:      feed.ID,
			Title:       item.Title,
			Description: desc,
			PublishedAt: t,
			Url:         feed.Url,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Printf("Failed to create post: %v", err)
		}
	}
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	req.Header.Add("User-Agent", "gator")
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("Failed to create a GET request: %v", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("Failed to send a GET request: %v", err)
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("Failed to read response body: %v", err)
	}
	var rssRes RSSFeed
	if err = xml.Unmarshal(bytes, &rssRes); err != nil {
		return &RSSFeed{}, fmt.Errorf("Failed to unmarshal response body: %v", err)
	}
	rssRes.Channel.Title = html.UnescapeString(rssRes.Channel.Title)
	rssRes.Channel.Description = html.UnescapeString(rssRes.Channel.Description)
	fmt.Println(rssRes)
	return &rssRes, nil
}
