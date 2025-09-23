package commands

import (
	"strings"
	"strconv"
	"context"
	"time"
	"log"
	"fmt"
	"github.com/google/uuid"
	"github.com/MedrekIT/gator/internal/config"
	"github.com/MedrekIT/gator/internal/aggregating"
	"github.com/MedrekIT/gator/internal/database"
)

func cmdHelp(s *config.State, cmd Command) error {
	fmt.Printf("Welcome to the Gator - your command line RSS feed aggregator!\n\nUsage:\n")
	for _, cmnd := range GetCommands() {
		fmt.Printf("'%s' - %s\n", cmnd.name, cmnd.description)
	}
	return nil
}

func cmdLogin(s *config.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'login <user_name>'\n")
	}

	user, err := s.Db.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			return fmt.Errorf("user with given name does not exist in the database\n")
		}
		return fmt.Errorf("error while getting user from the database - %w\n", err)
	}

	err = s.Conf.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User has been set to \"%s\"\n", user.Name)
	return nil
}

func cmdRegister(s *config.State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'register <user_name>'\n")
	}

	newUserParams := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.Args[0],
	}
	user, err := s.Db.CreateUser(context.Background(), newUserParams)
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"users_name_key\"") {
			return fmt.Errorf("user with given name already exists in the database\n")
		}
		return fmt.Errorf("error while creating new user - %w\n", err)
	}

	err = s.Conf.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("New user named \"%s\" has been created\n", user.Name)
	return nil
}

func cmdUsers(s *config.State, cmd Command) error {
	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error while getting users from the database\n")
	}

	if len(users) == 0 {
		fmt.Printf("No users in the database!\n")
	}
	for _, user := range users {
		if user.Name == s.Conf.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func cmdAddFeed(s *config.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("Incorrect usage\nTry 'addfeed <feed_name> <feed_url>'\n")
	}

	newFeedParams := database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.Args[0],
		Url: cmd.Args[1],
		UserID: user.ID,
	}
	feed, err := s.Db.CreateFeed(context.Background(), newFeedParams)
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"feeds_url_key\"") {
			return fmt.Errorf("feed with given URL already exists in the database\n")
		}
		return fmt.Errorf("error while creating new feed - %w\n", err)
	}

	newFeedFollowParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}
	feedFollow, err := s.Db.CreateFeedFollow(context.Background(), newFeedFollowParams)
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"feeds_url\"") {
			return fmt.Errorf("you already follow feed with given URL\n")
		}
		return fmt.Errorf("error while adding follow to the database - %w\n", err)
	}

	fmt.Printf("user \"%s\" now follows feed \"%s\"\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func cmdFeeds(s *config.State, cmd Command) error {
	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error while getting feeds from the database - %w\n", err)
	}

	if len(feeds) == 0 {
		fmt.Printf("No feeds in the database!\n")
	}
	for _, feed := range feeds {
		user, err := s.Db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			if strings.Contains(err.Error(), "sql: no rows in result set") {
				return fmt.Errorf("user with given ID does not exist in the database\n")
			}
			return fmt.Errorf("error while getting user from the database - %w\n", err)
		}

		fmt.Printf("\"%s\":\n", feed.Name)
		fmt.Printf(" * %s\n", feed.Url)
		fmt.Printf(" * %s\n", user.Name)
	}
	return nil
}

func cmdFollow(s *config.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'follow <feed_url>'\n")
	}

	feed, err := s.Db.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			return fmt.Errorf("feed with given URL does not exist in the database\n")
		}
		return fmt.Errorf("error while getting feed from the database - %w\n", err)
	}

	newFeedFollowParams := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}
	feedFollow, err := s.Db.CreateFeedFollow(context.Background(), newFeedFollowParams)
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"feed_follows_user_id_feed_id_key\"") {
			return fmt.Errorf("you already follow feed with given URL\n")
		}
		return fmt.Errorf("error while adding follow to the feed in the database - %w\n", err)
	}

	fmt.Printf("user \"%s\" now follows feed \"%s\"\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func cmdFollowing(s *config.State, cmd Command, user database.User) error {
	userFollows, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error while fetching follows data from the database - %w\n", err)
	}

	if len(userFollows) == 0 {
		fmt.Printf("You don't follow any feed!\n")
	}

	for _, feed := range userFollows {
		fmt.Printf("* \"%s\"\n", feed.FeedName)
	}
	return nil
}

func cmdUnfollow(s *config.State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'unfollow <feed_url>'\n")
	}

	feed, err := s.Db.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			return fmt.Errorf("given feed does not exist in the database\n")
		}
		return fmt.Errorf("error while getting feed from the database - %w\n", err)
	}

	newDeleteFollowParams := database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	err = s.Db.DeleteFeedFollow(context.Background(), newDeleteFollowParams)
	if err != nil {
		return fmt.Errorf("error while removing followed feed from the table - %w\n", err)
	}

	fmt.Printf("user \"%s\" now does not follow feed \"%s\"\n", user.Name, feed.Name)
	return nil
}

func cmdAgg(s *config.State, cmd Command) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("Incorrect usage\nTry 'agg <time_between_reqs [1s, 1m, 2h, 3m45s, ...(default = 1m)]>'\n")
	}

	var duration time.Duration
	var err error
	if len(cmd.Args) == 0 {
		duration, _ = time.ParseDuration("1m")
		fmt.Printf("Collecting feeds every 1m\n")
	} else {
		duration, err = time.ParseDuration(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("incorrect time format\nTry [1s, 1m, 2h, 3m45s, ...(default = 1m)]\n")
		}
		fmt.Printf("Collecting feeds every %s\n", cmd.Args[0])
	}

	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(duration)
	defer cancel()
	defer ticker.Stop()
	failures := 0
	for ;; <-ticker.C {
		select {
		case <-ctx.Done():
			log.Printf("Aggregating finished!\n")
			return nil
		case <-ticker.C:
			err := aggregating.ScrapeFeeds(s)
			if err != nil {
				failures++
				if failures >= 3 {
					log.Printf("error while scraping feeds data - %v\n", err)
					return fmt.Errorf("Too many consecutive errors, exiting...\n")
				}
				log.Printf("error while scraping feeds data - %v\nTrying again...\n", err)
				continue
			}
			failures = 0
		}
	}
}

func cmdBrowse(s *config.State, cmd Command, user database.User) error {
	if len(cmd.Args) > 1 {
		return fmt.Errorf("Incorrect usage\nTry 'browse <limit [default = 2]>'\n")
	}


	var postsLimit int
	var err error
	if len(cmd.Args) == 0 {
		postsLimit = 2
	} else {
		postsLimit, err = strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("incorrect limit format\nTry [1, 2, 3, ...(default = 2)]\n")
		}
	}

	newGetPostsParams := database.GetPostsForUserParams{
		ID: user.ID,
		Limit: int32(postsLimit),
	}
	posts, err := s.Db.GetPostsForUser(context.Background(), newGetPostsParams)
	if err != nil {
		return fmt.Errorf("error while fetching posts from the database - %v\n", err)
	}

	if len(posts) == 0 {
		fmt.Printf("There's nothing to browse!\n")
	}

	for _, post := range posts {
		fmt.Printf("\"%s\":\n", post.Title)
		fmt.Printf(" * %s\n", post.Description.String)
		fmt.Printf(" * %s\n\n", post.Url)
	}
	return nil
}

func cmdReset(s *config.State, cmd Command) error {
	err := s.Db.ResetDb(context.Background())
	if err != nil {
		return fmt.Errorf("error while resetting database - %w\n", err)
	}

	fmt.Printf("Database has been reset\n")
	return nil
}
