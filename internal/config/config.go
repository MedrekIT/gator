package config

import (
	"fmt"
	"os"
	"encoding/json"
)

const configFile = ".gatorconfig.json"

type Config struct {
	DbURL string `json:"db_url"`
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

func Read() (Config, error) {
	confPath, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}

	confBytes, err := os.ReadFile(confPath)
	if err != nil {
		return Config{}, fmt.Errorf("error fetching config file: %w\n", err)
	}

	var newConf Config
	if err := json.Unmarshal(confBytes, &newConf); err != nil {
		return Config{}, fmt.Errorf("error decoding data from config file: %w\n", err)
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
		return fmt.Errorf("error encoding config data: %w\n", err)
	}

	err = os.WriteFile(confPath, confBytes, 0666)
	if err != nil {
		return fmt.Errorf("error saving data to file: %w\n", err)
	}

	return nil
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting HOME_DIR location: %w\n", err)
	}

	return fmt.Sprintf("%s/%s", homeDir, configFile), nil
}
