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
	InitialActive bool          `json:"initialHelloActive"`
	Filter        Filters       `json:"filters"`
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

type Filters struct {
	Caps CapsFilter `json:"caps"`
	Word Words      `json:"words"`
	Max  MaxLength  `json:"maxLength"`
}

type CapsFilter struct {
	Min     int     `json:"min"`
	Percent float64 `json:"percent"`
	Active  bool    `json:"active"`
	Penalty Penalty `json:"penalty"`
	Message string  `json:"message"`
}

type Penalty struct {
	Type     string `json:"type"`
	Duration int    `json:"duration"`
}

type Words struct {
	Active  bool       `json:"active"`
	BanList []BanWords `json:"banLists"`
}

type BanWords struct {
	Words   []string `json:"words"`
	Penalty Penalty  `json:"penalty"`
	Message string   `json:"message"`
}

type MaxLength struct {
	Active  bool    `json:"active"`
	Max     int     `json:"max"`
	Message string  `json:"message"`
	Penalty Penalty `json:"penalty"`
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
