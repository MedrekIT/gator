package main

import (
	"os"
	"log"
	"database/sql"
	"github.com/MedrekIT/gator/internal/config"
	"github.com/MedrekIT/gator/internal/database"
)

import _ "github.com/lib/pq"

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatalf("\nError - %v", err)
	}

	db, err := sql.Open("postgres", conf.DbURL)
	dbQueries := database.New(db)

	if len(os.Args) < 2 {
		log.Fatalf("\nError - Usage: cli <command> [args...]")
	}

	s := state{dbQueries, &conf}
	cmds := commands{make(map[string]func(s *state, cmd command) error)}
	cmds.register("register", cmdRegister)
	cmds.register("login", cmdLogin)
	cmds.register("users", cmdUsers)
	cmds.register("reset", cmdReset)

	cmd := command{os.Args[1], os.Args[2:]}

	err = cmds.run(&s, cmd)
	if err != nil {
		log.Fatalf("\nError - %v", err)
	}
}
