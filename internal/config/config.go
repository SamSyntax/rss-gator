package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const fileName = ".gatorconfig.json"

type Config struct {
	DB_URL            string `json:"db_url"`
	CURRENT_USER_NAME string `json:"current_user_name"`
}

func Read() Config {
	cfg := Config{}
	hmdir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Failed to read homedir")
	}
	path := fmt.Sprintf("%s/.gatorconfig.json", hmdir)
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Failed to open file")
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Println("Failed to decode json", err)
	}
	return cfg
}

func SetUser(user string) error {
	cfg := Read()
	cfg.CURRENT_USER_NAME = user
	return Write(cfg)
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Failed to get a home directory: %v", err)
	}
	return home + "/" + fileName, nil
}

func Write(cfg Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("Failed to marshal config: %v", err)
	}
	err = os.WriteFile(path, bytes, 0o644)
	if err != nil {
		return fmt.Errorf("Failed write config file: %v", err)
	}
	return nil
}
