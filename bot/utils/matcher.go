package utils

import (
	"log"
	"regexp"
)

type Matcher struct {
	rgxp []*regexp.Regexp
}

func NewMatcher(words []string, log *log.Logger) Matcher {
	m := Matcher{}
	for _, w := range words {
		r, err := regexp.Compile("(.*\\W|^)" + w + "(\\W.*|$)")
		if err != nil {
			log.Println("Cant compile expression for filter word: " + w)
		} else {
			m.rgxp = append(m.rgxp, r)
		}
	}
	return m
}

func (m *Matcher) Match(s string) bool {
	for i := range m.rgxp {
		if m.rgxp[i].MatchString(s) {
			return true
		}
	}
	return false
}
