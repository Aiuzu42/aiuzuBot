package utils

import (
	"math/rand"
	"time"
	"unicode"
)

const (
	Version = "3.1.0"
)

var actions = []string{"response"}
var penalties = []string{"temporary", "permanent", ""}
var username = make(map[string]string)

//ValidateResponseType returns true if the parameter t is one of the valid action types.
//Returns false otherwise.
func ValidateResponseType(t string) bool {
	return ExistsInSlice(t, actions)
}

//ExistsInSlice receives an string and an string slice.
//Returns true if the string is an elemnt of the slice.
//Returns false otherwise.
func ExistsInSlice(s string, sl []string) bool {
	for _, e := range sl {
		if s == e {
			return true
		}
	}
	return false
}

//AddToUsers adds the userId and name of a user to the username table.
func AddToUsers(userId string, name string) {
	username[userId] = name
}

//GetUserName looks for the userId name in the username table.
//Uses time.Now().Unix() as seed.
func GetUserName(userId string) string {
	return username[userId]
}

//GetRandomElement returns a random element from the provided slice.
func GetRandomElement(s []string) string {
	rand.Seed(time.Now().Unix())
	return s[rand.Intn(len(s))]
}

func ValidateCaps(p float64, min int, msg string) bool {
	b := true
	if len(msg) >= min {
		c := 0
		l := int(float64(len(msg)) * p)
		for _, r := range msg {
			if unicode.IsLetter(r) && unicode.IsUpper(r) {
				c++
			}
			if c >= l {
				b = false
				break
			}
		}
	}
	return b
}

func MassMatchWords(w []string, s string) bool {
	for _, word := range w {
		if MatchWord(word, s) {
			return true
		}
	}
	return false
}

func MatchWord(w string, s string) bool {
	return false
}
