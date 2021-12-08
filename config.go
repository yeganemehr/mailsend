package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	AllowTo []string `json:"allow-to"`
}

func ParseConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
