package main

import (
	"strings"
	"context"
	"fmt"
	"os"
	"log"
	"database/sql"
	"github.com/MedrekIT/gator/internal/config"
	"github.com/MedrekIT/gator/internal/database"
)

import _ "github.com/lib/pq"

func middlewareLoggedIn(handler func(s *config.State, cmd command, user database.User) error) func(*config.State, command) error {
	return func(s *config.State, cmd command) error {
		user, err := s.Db.GetUser(context.Background(), s.Conf.CurrentUserName)
		if err != nil {
			if strings.Contains(err.Error(), "sql: no rows in result set") {
				return fmt.Errorf("user with given name doesn't exist in the database\n")
			}
			return fmt.Errorf("error while getting user from the database - %w\n", err)
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

	s := config.State{
		Db: dbQueries,
		Conf: &conf,
	}
	cmds := getCommands()

	cmd := command{os.Args[1], os.Args[2:]}

	err = cmds[os.Args[1]].run(&s, cmd)
	if err != nil {
		log.Fatalf("\nError: %v", err)
	}
}
