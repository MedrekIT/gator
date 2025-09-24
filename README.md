# Gator - RSS feed aggregator

Gator - RSS feed aggregator which allows you to follow to your favorite blogs and podcast if they are available in RSS.

---

## Table of Contents

- [Screenshots](#screenshots)
- [Motivation](#motivation)
- [Information](#information)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)

---

## Screenshots

![Browse](./screenshots/gator_browse.png)

---

## Motivation

In modern entertainment and research you are pretty much always being controlled by "algorithms" they know the best what is good for you, and choose what you have to see, of course, you may find something valuable online, but the risk is that this mysterious algorithm will decide that is is not worth informing you about related news or posts, then you will probably forget what it was and that it exists.

You can prevent it by simply using an RSS aggregator, where YOU decide what is worth your time and what you want to read/listen about.

---

## Information

> [!NOTE]
> **Requirements:**
> - `go>=1.25.1`
>
> A config file called `.gatorconfig.json` as well as `db/.gator.db` are created in `HOME_DIR/.local/share/gator/` directory with preconfigured database path.

---

## Installation

```bash
curl -sS https://webi.sh/golang | sh # Install Go
go install github.com/MedrekIT/gator@latest # Install repository as program for global execution
```

---

## Usage

### Run
```bash
gator <command> [...args]
```

You are able to add some users (which you may treat as profiles/context groups), once you login, you may add any RSS feeds you'd like. Run `gator agg <time_interval>` in unused terminal, then try to browse your posts.

Here are some good blogs for you to start with your RSS adventure:
- `gator addfeed "Boot.dev Blog" https://blog.boot.dev/index.xml`
- `gator addfeed "Frontend Masters Blog" https://frontendmasters.com/blog/feed`

**Commands:**
- `gator help` - Displays all implemented commands
- `gator register <user_name>` - Allows to register new user account
- `gator login <user_name>` - Allows registered user to login onto existant account
- `gator users` - Displays all registered users
- `gator addfeed <feed_name> <feed_url>` - Allows to save and follow a new RSS feed
- `gator feeds` - Displays all feeds saved by users
- `gator follow <feed_url>` - Allows to follow any saved feed
- `gator unfollow <feed_url>` - Allows to unfollow any followed feed
- `gator following` - Displays every feed that you follow
- `gator agg <time_between_reqs [1s, 1m, 2h, 3m45s, ...]>` - Starts the automatic feeds aggregation and fetches new posts whenever given time passes
- `gator browse <limit [default = 2]> <feed_query>` - Displays number of freshly fetched posts for current user, limited by given value, may be filtered by specified feed name's part
- `gator reset` - Resets all saved data

---

## Contributing

### Install project

```bash
git clone https://github.com/MedrekIT/gator.git # Clone repository
cd gator # Move to repository directory
```

### Test if it works

```bash
go run . help
```

### Add you changes and submit a pull request

If you'd like to contribute, please fork the repository and open a pull request to the `main` branch.
