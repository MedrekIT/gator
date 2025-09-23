package commands

import (
	"context"
	"strings"
	"fmt"
	"github.com/MedrekIT/gator/internal/database"
	"github.com/MedrekIT/gator/internal/config"
)

func middlewareLoggedIn(handler func(s *config.State, cmd Command, user database.User) error) func(*config.State, Command) error {
	return func(s *config.State, cmd Command) error {
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


