package main

import (
	"context"
	"fmt"
	"os"
	"log"
	"database/sql"
	"github.com/MedrekIT/gator/internal/config"
	"github.com/MedrekIT/gator/internal/database"
)

import _ "github.com/lib/pq"

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.conf.CurrentUserName)
		if err != nil {
			return fmt.Errorf("user with given name doesn't exist in the database\n")
		}

		return handler(s, cmd, user)
	}
}

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatalf("\nError - %v", err)
	}

	db, err := sql.Open("postgres", conf.DbURL)
	dbQueries := database.New(db)

	if len(os.Args) < 2 {
		log.Fatalf("\nUsage: cli <command> [args...]")
	}

	s := state{dbQueries, &conf}
	cmds := commands{make(map[string]func(s *state, cmd command) error)}
	cmds.register("register", cmdRegister)
	cmds.register("login", cmdLogin)
	cmds.register("users", cmdUsers)
	cmds.register("addfeed", middlewareLoggedIn(cmdAddFeed))
	cmds.register("feeds", cmdFeeds)
	cmds.register("follow", middlewareLoggedIn(cmdFollow))
	cmds.register("following", middlewareLoggedIn(cmdFollowing))
	cmds.register("unfollow", middlewareLoggedIn(cmdUnfollow))
	cmds.register("agg", cmdAgg)
	cmds.register("reset", cmdReset)

	cmd := command{os.Args[1], os.Args[2:]}

	err = cmds.run(&s, cmd)
	if err != nil {
		log.Fatalf("\nError - %v", err)
	}
}
