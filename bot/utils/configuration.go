package utils

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type Config struct {
	Configuration Configuration `json:"configuration"`
	Actions       []Action      `json:"actions"`
	Quotes        []string      `json:"quotes"`
	InitialHello  string        `json:"initialHello"`
}

type Configuration struct {
	ApiKey              string   `json:"apiKey"`
	Refresh             string   `json:"refresh"`
	ClientId            string   `json:"clientId"`
	ClientS             string   `json:"clientS"`
	LiveStreamChannelId string   `json:"liveStreamChannelId"`
	AuthorId            string   `json:"authorId"`
	Admins              []string `json:"admins"`
	Excluded            []string `json:"excluded"`
}

type Action struct {
	Name          string   `json:"name"`
	Keywords      []string `json:"keywords"`
	Type          string   `json:"type"`
	Message       string   `json:"message"`
	UserTimeout   int64    `json:"userTimeout"`
	GlobalTimeout int64    `json:"globalTimeout"`
	Admin         bool     `json:"admin"`
}

func LoadConfig() (Config, error) {
	var c Config
	file, err := os.Open("configurations.json")
	if err != nil {
		log.Println("Unable to load configuration file.")
		return Config{}, errors.New("Unable to load configuration file")
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&c)
	return c, nil
}
