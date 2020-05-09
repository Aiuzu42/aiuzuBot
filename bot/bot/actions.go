package bot

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

var ErrNotAuthorized = errors.New("Not authorized to run that command.")
var ErrActionTypeNotFound = errors.New("The action type is not valid.")

type Action struct {
	Name          string
	Keywords      []string
	Type          string
	Message       string
	UserTimeout   int64
	GlobalTimeout int64
	Admin         bool
	LastCalled    int64
	UserList      map[string]int64
}

//reaminingTimeout calculates how many seconds until an action can be called again.
//A positive return value is the number of seconds until the action is available again.
//A zero or negative value means the action can be called.
func remainingTimeout(now int64, timeout int64, last int64) int64 {
	if timeout > 0 {
		return (last + timeout) - now
	} else {
		return 0
	}
}

func (a *Action) validateTimeout(user string, now int64, l *log.Logger) error {
	global := remainingTimeout(now, a.GlobalTimeout, a.LastCalled)
	if global > 0 {
		e := fmt.Sprintf("%d seconds remaining to use %s command again[global].", global, a.Name)
		l.Println(e)
		return errors.New(e)
	}
	ut := a.UserList[user]
	timeUser := remainingTimeout(now, a.UserTimeout, ut)
	if timeUser > 0 {
		e := fmt.Sprintf("%s has %d seconds remaining to use %s command again[user].", user, global, a.Name)
		l.Println(e)
		return errors.New(e)
	}
	return nil
}

func (a *Action) findKeyword(msg string) bool {
	for _, k := range a.Keywords {
		if strings.Contains(msg, k) {
			return true
		}
	}
	return false
}
