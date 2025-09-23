package main

import (
	"os"
	"log"
	"database/sql"
	"github.com/MedrekIT/gator/internal/database"
	"github.com/MedrekIT/gator/internal/config"
	"github.com/MedrekIT/gator/internal/commands"
)

import _ "github.com/lib/pq"

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatalf("\nError: %v", err)
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
	cmds := commands.GetCommands()

	cmd := commands.Command{os.Args[1], os.Args[2:]}

	if _, ok := cmds[os.Args[1]]; !ok {
		log.Fatalf("\nCommand '%s' not specified\nTry 'help' to see all commands\n", os.Args[1])
	}
	err = cmds[os.Args[1]].Run(&s, cmd)
	if err != nil {
		log.Fatalf("\nError: %v", err)
	}
}
