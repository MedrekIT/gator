package commands

import (
	"github.com/MedrekIT/gator/internal/config"
)

func GetCommands() map[string]commands {
	return map[string]commands{
		"help": {
			name: "help",
			callback: cmdHelp,
			description: "Displays this help message",
		}, "register": {
			name: "register <user_name>",
			callback: cmdRegister,
			description: "Allows to register new user account",
		}, "login": {
			name: "login <user_name>",
			callback: cmdLogin,
			description: "Allows registered user to login onto existant account",
		}, "users": {
			name: "users",
			callback: cmdUsers,
			description: "Displays all registered users",
		}, "addfeed": {
			name: "addfeed <feed_name> <feed_url>",
			callback: middlewareLoggedIn(cmdAddFeed),
			description: "Allows to save and follow a new RSS feed",
		}, "feeds": {
			name: "feeds",
			callback: cmdFeeds,
			description: "Displays all feeds saved by users",
		}, "follow": {
			name: "follow <feed_url>",
			callback: middlewareLoggedIn(cmdFollow),
			description: "Allows to follow any saved feed",
		}, "unfollow": {
			name: "unfollow <feed_url>",
			callback: middlewareLoggedIn(cmdUnfollow),
			description: "Allows to unfollow any followed feed",
		}, "following": {
			name: "following",
			callback: middlewareLoggedIn(cmdFollowing),
			description: "Displays every feed that you follow",
		}, "agg": {
			name: "agg <time_between_reqs [1s, 1m, 2h, 3m45s, ...(default = 1m)]>",
			callback: cmdAgg,
			description: "Starts the automatic feeds aggregation and fetches new posts whenever given time passes",
		}, "browse": {
			name: "browse <limit [default = 2]>",
			callback: middlewareLoggedIn(cmdBrowse),
			description: "Displays number of freshly fetched posts limited by given value",
		}, "reset": {
			name: "reset",
			callback: cmdReset,
			description: "Resets all saved data",
		},
	}
}

type commands struct {
	name string
	callback func(*config.State, Command) error
	description string
}

type Command struct {
	Name string
	Args []string
}

func (c commands) Run(s *config.State, cmd Command) error {
	err := c.callback(s, cmd)
	if err != nil {
		return err
	}

	return nil
}
