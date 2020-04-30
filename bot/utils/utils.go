package utils

import (
	"math/rand"
	"time"
)

const (
	Version = "1.1.1"
)

var Game = ""

var admins []string
var actions = []string{"response", "setupgame"}
var users = make(map[string]string)
var quotes = []string{}
var ThisBot string
var InitialHello string

func SetAdmins(a []string) {
	admins = a
}

func IsAdmin(u string) bool {
	return existsInSlice(u, admins)
}

func ValidateResponseType(t string) bool {
	return existsInSlice(t, actions)
}

func existsInSlice(s string, sl []string) bool {
	for _, e := range sl {
		if s == e {
			return true
		}
	}
	return false
}
func AddToUsers(uid string, name string) {
	users[uid] = name
}

func GetUserName(uid string) string {
	return users[uid]
}

func GetRandomQuote() string {
	rand.Seed(time.Now().Unix())
	return quotes[rand.Intn(len(quotes))]
}

func SetQuotes(q []string) {
	quotes = q
}

func IsExcluded(userId string) bool {
	if userId == ThisBot {
		return true
	} else {
		return false
	}
}
