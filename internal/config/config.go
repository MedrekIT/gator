package config

import (
	"fmt"
	"os"
	"encoding/json"
	"github.com/MedrekIT/gator/internal/database"
)

const configPath = "/.local/share/gator"
const configFile = ".gatorconfig.json"
const dbPath = "/.local/share/gator/db/"
const dbFile = "gator.db"

type State struct {
	Db *database.Queries
	Conf *Config
}

type Config struct {
	DbPath string `json:"db_path"`
	CurrentUserName string `json:"current_user_name"`
}

func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName

	err := write(c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) EnsureDBPath() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error while getting HOME_DIR location - %w\n", err)
	}

	dir := homeDir + dbPath
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error while creating database path - %w\n", err)
	}

	c.DbPath = dir + dbFile
	return nil
}

func Read() (Config, error) {
	confPath, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}

	confBytes, err := os.ReadFile(confPath)
	if err != nil {
		conf := Config{
			DbPath: dbPath,
		}
		err = conf.EnsureDBPath()
		if err != nil {
			return Config{}, err
		}
		confBytes, err = json.Marshal(conf)
		if err != nil {
			return Config{}, fmt.Errorf("error while encoding config data - %w\n", err)
		}
		err = os.WriteFile(confPath, confBytes, 0666)
		if err != nil {
			return Config{}, fmt.Errorf("error while creating non-existant config file - %w\n", err)
		}
	}

	var newConf Config
	if err := json.Unmarshal(confBytes, &newConf); err != nil {
		return Config{}, fmt.Errorf("error while decoding data from config file - %w\n", err)
	}

	return newConf, nil
}

func write(conf *Config) error {
	confPath, err := getConfigPath()
	if err != nil {
		return err
	}

	confBytes, err := json.Marshal(conf)
	if err != nil {
		return fmt.Errorf("error while encoding config data - %w\n", err)
	}

	err = os.WriteFile(confPath, confBytes, 0666)
	if err != nil {
		return fmt.Errorf("error while saving data to file - %w\n", err)
	}

	return nil
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error while getting HOME_DIR location - %w\n", err)
	}

	dir := homeDir + configPath
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("error while creating config path - %w\n", err)
	}

	return fmt.Sprintf("%s/%s", dir, configFile), nil
}
