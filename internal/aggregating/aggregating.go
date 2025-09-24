package aggregating

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"io"
	"net/http"
	"encoding/xml"
	"html"
	"github.com/google/uuid"
	"github.com/MedrekIT/gator/internal/config"
	"github.com/MedrekIT/gator/internal/database"
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

func ScrapeFeeds(s *config.State) error {
	feed, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			return fmt.Errorf("nothing to browse, database is empty!\n")
		}
		return fmt.Errorf("error while getting feeds to fetch from the database - %w\n", err)
	}

	newMarkFeedParams := database.MarkFeedFetchedParams{
		ID: feed.ID,
		UpdatedAt: time.Now(),
	}
	_, err = s.Db.MarkFeedFetched(context.Background(), newMarkFeedParams)
	if err != nil {
		return fmt.Errorf("error while marking feed as fetched in the database - %w\n", err)
	}

	fetchedFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, it := range fetchedFeed.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, it.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}

		newPostParams := database.CreatePostParams{
			ID: uuid.New().String(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title: it.Title,
			Url: it.Link,
			Description: sql.NullString{
				String: it.Description,
				Valid: true,
			},
			PublishedAt: publishedAt,
			FeedID: feed.ID,
		}
		_, err = s.Db.CreatePost(context.Background(), newPostParams)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: posts.url") {
				continue
			}
			return fmt.Errorf("error while aggregating feed posts - %w\n", err)
		}
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error while creating request - %w\n", err)
	}

	req.Header.Set("user-agent", "gator")
	req.Header.Set("content-type", "application/xml")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error while fetching response - %w\n", err)
	}

	if res.StatusCode > 299 {
		return &RSSFeed{}, fmt.Errorf("response failed with status - %s\n", res.Status)
	}
	defer res.Body.Close()

	dataBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error while reading data from body - %w\n", err)
	}

	var data RSSFeed
	if err := xml.Unmarshal(dataBytes, &data); err != nil {
		return &RSSFeed{}, fmt.Errorf("error while decoding response body - %w\n", err)
	}

	data.Channel.Title = html.UnescapeString(data.Channel.Title)
	data.Channel.Description = html.UnescapeString(data.Channel.Description)
	for i, it := range data.Channel.Item {
		data.Channel.Item[i].Title = html.UnescapeString(it.Title)
		data.Channel.Item[i].Description = html.UnescapeString(it.Description)
	}

	return &data, nil
}
