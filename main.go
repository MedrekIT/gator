package main

import (
	"os"
	"log"
	"database/sql"
	"github.com/MedrekIT/gator/internal/database"
	"github.com/MedrekIT/gator/internal/config"
	"github.com/MedrekIT/gator/internal/commands"
	"github.com/MedrekIT/gator/sql"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatalf("\nError: %v", err)
	}

	db, err := sql.Open("sqlite3", conf.DbPath)
	if err != nil {
		log.Fatalf("\nError: %v", err)
	}
	defer db.Close()

	err = embedding.DbEmbedding(db)
	if err != nil {
		log.Fatalf("\nError: %v", err)
	}

	dbQueries := database.New(db)

	if len(os.Args) < 2 {
		log.Fatalf("\nUsage: cli <command> [args...]")
	}

	s := config.State{
		Db: dbQueries,
		Conf: &conf,
	}
	cmds := commands.GetCommands()

	cmd := commands.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	if _, ok := cmds[os.Args[1]]; !ok {
		log.Fatalf("\nCommand '%s' not specified\nTry 'help' to see all commands\n", os.Args[1])
	}
	err = cmds[os.Args[1]].Run(&s, cmd)
	if err != nil {
		log.Fatalf("\nError: %v", err)
	}
}
