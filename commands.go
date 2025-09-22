package main

import (
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

type commands struct {
	cmds map[string]func(*config.State, command) error
}

type command struct {
	name string
	args []string
}

func (c *commands) run(s *config.State, cmd command) error {
	err := c.cmds[cmd.name](s, cmd)
	if err != nil {
		return err
	}

	return nil
}

func (c *commands) register(name string, f func(*config.State, command) error) {
	c.cmds[name] = f
}

func cmdLogin(s *config.State, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'login <user_name>'\n")
	}

	user, err := s.Db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("user with given name does not exist in the database\n")
	}

	err = s.Conf.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User has been set to \"%s\"\n", user.Name)
	return nil
}

func cmdRegister(s *config.State, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'register <user_name>'\n")
	}

	newUserParams := database.CreateUserParams{uuid.New(), time.Now(), time.Now(), cmd.args[0]}
	user, err := s.Db.CreateUser(context.Background(), newUserParams)
	if err != nil {
		return fmt.Errorf("user with given name already exists in the database\n")
	}

	err = s.Conf.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("New user named \"%s\" has been created\n", user.Name)
	return nil
}

func cmdUsers(s *config.State, cmd command) error {
	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("no users in the database\n")
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

func cmdAddFeed(s *config.State, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("Incorrect usage\nTry 'addfeed <feed_name> <feed_url>'\n")
	}

	newFeedParams := database.CreateFeedParams{uuid.New(), time.Now(), time.Now(), cmd.args[0], cmd.args[1], user.ID}
	feed, err := s.Db.CreateFeed(context.Background(), newFeedParams)
	if err != nil {
		return fmt.Errorf("feed with given URL already exists in the database\n")
	}

	newFeedFollowParams := database.CreateFeedFollowParams{uuid.New(), time.Now(), time.Now(), user.ID, feed.ID}
	feedFollow, err := s.Db.CreateFeedFollow(context.Background(), newFeedFollowParams)
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

func cmdFeeds(s *config.State, cmd command) error {
	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("no feeds in the database\n")
	}

	for _, feed := range feeds {
		user, err := s.Db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("user with given ID does not exist in the database\n")
		}

		fmt.Printf("\"%s\":\n", feed.Name)
		fmt.Printf(" * %s\n", feed.Url)
		fmt.Printf(" * %s\n", user.Name)
	}
	return nil
}

func cmdFollow(s *config.State, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'follow <feed_url>'\n")
	}

	feed, err := s.Db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("feed with given URL does not exist in the database\n")
	}

	newFeedFollowParams := database.CreateFeedFollowParams{uuid.New(), time.Now(), time.Now(), user.ID, feed.ID}
	feedFollow, err := s.Db.CreateFeedFollow(context.Background(), newFeedFollowParams)
	if err != nil {
		return fmt.Errorf("you already follow feed with given URL\n")
	}

	fmt.Printf("user \"%s\" now follows feed \"%s\"\n", feedFollow.UserName, feedFollow.FeedName)
	return nil
}

func cmdFollowing(s *config.State, cmd command, user database.User) error {
	userFollows, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
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

func cmdUnfollow(s *config.State, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'unfollow <feed_url>'\n")
	}

	feed, err := s.Db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("given feed does not exist in the database\n")
	}

	newDeleteFollowParams := database.DeleteFeedFollowParams{user.ID, feed.ID}
	_, err = s.Db.DeleteFeedFollow(context.Background(), newDeleteFollowParams)
	if err != nil {
		return fmt.Errorf("you were not following that feed\n")
	}

	fmt.Printf("user \"%s\" now does not follow feed \"%s\"\n", user.Name, feed.Name)
	return nil
}

func cmdAgg(s *config.State, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'agg <time_between_reqs [1s, 1m, 2h, 3m45s, ...]>'\n")
	}

	duration, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("incorrect time format\n")
	}

	fmt.Printf("Collecting feeds every %s\n", cmd.args[0])

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
					log.Printf("error while scraping feeds data - %w\n", err)
					return fmt.Errorf("Too many consecutive errors, exiting...\n")
				}
				log.Printf("error while scraping feeds data - %w\nTrying again...\n", err)
				continue
			}
			failures = 0
		}
	}
	return nil
}

func cmdBrowse(s *config.State, cmd command, user database.User) error {
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

	posts, err := s.Db.GetPostsForUser(context.Background(), newGetPostsParams)
	if err != nil {
		return fmt.Errorf("error while fetching posts from the database - %v\n", err)
	}

	for _, post := range posts {
		fmt.Printf("\"%s\":\n", post.Title)
		fmt.Printf(" * %s\n", post.Description.String)
		fmt.Printf(" * %s\n\n", post.Url)
	}
	return nil
}

func cmdReset(s *config.State, cmd command) error {
	_, err := s.Db.ResetDb(context.Background())
	if err != nil {
		return fmt.Errorf("error while resetting database - %w\n", err)
	}

	fmt.Printf("Database has been reset\n")
	return nil
}
