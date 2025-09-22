package main

import (
	"strconv"
	"context"
	"time"
	"fmt"
	"io"
	"net/http"
	"encoding/xml"
	"html"
	"database/sql"
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

type commands struct {
	cmds map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	err := c.cmds[cmd.name](s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

type command struct {
	name string
	args []string
}

type state struct {
	db *database.Queries
	conf *config.Config
}

func cmdLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'login <user_name>'\n")
	}

	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("user with given name does not exist in the database\n")
	}

	err = s.conf.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User has been set to \"%s\"\n", user.Name)
	return nil
}

func cmdRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'register <user_name>'\n")
	}

	newUserParams := database.CreateUserParams{uuid.New(), time.Now(), time.Now(), cmd.args[0]}
	user, err := s.db.CreateUser(context.Background(), newUserParams)
	if err != nil {
		return fmt.Errorf("user with given name already exists in the database\n")
	}

	err = s.conf.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("New user named \"%s\" has been created\n", user.Name)
	return nil
}

func cmdUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("no users in the database\n")
	}

	for _, user := range users {
		if user.Name == s.conf.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func cmdAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("Incorrect usage\nTry 'addfeed <feed_name> <feed_url>'\n")
	}

	newFeedParams := database.CreateFeedParams{uuid.New(), time.Now(), time.Now(), cmd.args[0], cmd.args[1], user.ID}
	feed, err := s.db.CreateFeed(context.Background(), newFeedParams)
	if err != nil {
		return fmt.Errorf("feed with given URL already exists in the database\n")
	}

	newFeedFollowParams := database.CreateFeedFollowParams{uuid.New(), time.Now(), time.Now(), user.ID, feed.ID}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), newFeedFollowParams)
	if err != nil {
		return fmt.Errorf("you already follow feed with given URL\n")
	}

	fmt.Printf("user \"%s\" now follows feed \"%s\"\n", feedFollow.UserName, feedFollow.FeedName)

	fmt.Println(feed.ID)
	fmt.Println(feed.CreatedAt)
	fmt.Println(feed.UpdatedAt)
	fmt.Println(feed.Name)
	fmt.Println(feed.Url)
	fmt.Println(feed.UserID)
	return nil
}

func cmdFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("no feeds in the database\n")
	}

	for _, feed := range feeds {
		user, err := s.db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("user with given ID does not exist in the database\n")
		}

		fmt.Printf("\"%s\":\n", feed.Name)
		fmt.Printf(" * %s\n", feed.Url)
		fmt.Printf(" * %s\n", user.Name)
	}
	return nil
}

func cmdFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'follow <feed_url>'\n")
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("feed with given URL does not exist in the database\n")
	}

	newFeedFollowParams := database.CreateFeedFollowParams{uuid.New(), time.Now(), time.Now(), user.ID, feed.ID}
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), newFeedFollowParams)
	if err != nil {
		return fmt.Errorf("you already follow feed with given URL\n")
	}

	fmt.Printf("user \"%s\" now follows feed \"%s\"\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func cmdFollowing(s *state, cmd command, user database.User) error {
	userFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error while fetching follows data from the database - %w\n", err)
	}

	if len(userFollows) == 0 {
		fmt.Printf("you don't follow any feed\n")
	}

	for _, feed := range userFollows {
		fmt.Printf("* \"%s\"\n", feed.FeedName)
	}
	return nil
}

func cmdUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'unfollow <feed_url>'\n")
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("given feed does not exist in the database\n")
	}

	newDeleteFollowParams := database.DeleteFeedFollowParams{user.ID, feed.ID}
	_, err = s.db.DeleteFeedFollow(context.Background(), newDeleteFollowParams)
	if err != nil {
		return fmt.Errorf("you were not following that feed\n")
	}

	fmt.Printf("user \"%s\" now does not follow feed \"%s\"\n", user.Name, feed.Name)
	return nil
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("you have no new feeds to fetch\n")
	}
	fmt.Printf("test1")

	newMarkFeedParams := database.MarkFeedFetchedParams{feed.ID, time.Now()}
	_, err = s.db.MarkFeedFetched(context.Background(), newMarkFeedParams)
	if err != nil {
		return err
	}
	fmt.Printf("test2")

	fetchedFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}
	fmt.Printf("test3")

	for _, it := range fetchedFeed.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, it.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}

		newPostParams := database.CreatePostParams{uuid.New(), time.Now(), time.Now(), it.Title, it.Link, sql.NullString{it.Description, true}, publishedAt, feed.ID}
		_, err = s.db.CreatePost(context.Background(), newPostParams)
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error creating request - %w\n", err)
	}

	req.Header.Set("user-agent", "gator")
	req.Header.Set("content-type", "application/xml")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error fetching response - %w\n", err)
	}

	if res.StatusCode > 299 {
		return &RSSFeed{}, fmt.Errorf("response failed with status - %s\n", res.Status)
	}
	defer res.Body.Close()

	dataBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error reading data from body - %w\n", err)
	}

	var data RSSFeed
	if err := xml.Unmarshal(dataBytes, &data); err != nil {
		return &RSSFeed{}, fmt.Errorf("error decoding response body - %w\n", err)
	}

	data.Channel.Title = html.UnescapeString(data.Channel.Title)
	data.Channel.Description = html.UnescapeString(data.Channel.Description)
	for i, it := range data.Channel.Item {
		data.Channel.Item[i].Title = html.UnescapeString(it.Title)
		data.Channel.Item[i].Description = html.UnescapeString(it.Description)
	}

	return &data, nil
}

func cmdAgg(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'agg <time_between_reqs [1s, 1m, 2h, 3m45s, ...]>'\n")
	}

	duration, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("incorrect time format\n")
	}

	fmt.Printf("Collecting feeds every %s\n", cmd.args[0])
	ticker := time.NewTicker(duration)
	for ;; <-ticker.C {
		scrapeFeeds(s)
	}
	return nil
}

func cmdBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 1 {
		return fmt.Errorf("Incorrect usage\nTry 'browse <limit [default = 2]>'\n")
	}


	newGetPostsParams := database.GetPostsForUserParams{}
	if len(cmd.args) == 0 {
		newGetPostsParams = database.GetPostsForUserParams{user.ID, 2}
	} else {
		postsLimit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit format\n")
		}
		newGetPostsParams = database.GetPostsForUserParams{user.ID, int32(postsLimit)}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), newGetPostsParams)
	if err != nil {
		return fmt.Errorf("error while fetching posts from the database - %w\n", err)
	}

	for _, post := range posts {
		fmt.Printf("%s:\n", post.Title)
		fmt.Printf(" * %s\n", post.Description.String)
		fmt.Printf(" * %s\n\n", post.Url)
	}
	return nil
}

func cmdReset(s *state, cmd command) error {
	_, err := s.db.ResetDb(context.Background())
	if err != nil {
		return fmt.Errorf("error while resetting database - %w\n", err)
	}

	fmt.Printf("Database has been reset\n")
	return nil
}
