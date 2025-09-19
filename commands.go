package main

import (
	"context"
	"time"
	"fmt"
	"github.com/google/uuid"
	"github.com/MedrekIT/gator/internal/config"
	"github.com/MedrekIT/gator/internal/database"
)

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
		return fmt.Errorf("user with given name doesn't exist in the database\n")
	}

	err = s.conf.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User has been set to %s\n", user.Name)
	return nil
}

func cmdRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Incorrect usage\nTry 'register <user_name>'\n")
	}


	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{uuid.New(), time.Now(), time.Now(), cmd.args[0]})
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
		return fmt.Errorf("no users in the database")
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

func cmdReset(s *state, cmd command) error {
	_, err := s.db.ResetDb(context.Background())
	if err != nil {
		return fmt.Errorf("error while resetting database: %w\n", err)
	}

	fmt.Printf("Database has been reset")
	return nil
}
