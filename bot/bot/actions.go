package bot

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aiuzu42/aiuzuBot/bot/utils"
	"github.com/aiuzu42/aiuzuBot/bot/youtubeapi"
)

var ErrNotAuthorized = errors.New("Not authorized to run that command.")
var ErrActionTypeNotFound = errors.New("The action type is not valid.")

type Action struct {
	Name          string   `json:"name"`
	Keywords      []string `json:"keywords"`
	Type          string   `json:"type"`
	Message       string   `json:"message"`
	UserTimeout   int64    `json:"userTimeout"`
	GlobalTimeout int64    `json:"globalTimeout"`
	Admin         bool     `json:"admin"`

	//The last time the action was executed
	LastCalled int64

	//A list of userId and the last time the action was executed by them
	UserList map[string]int64
}

func (a *Action) ExecuteAction(userId string, chat string, author string, msg string, l *log.Logger) error {
	//First we validate if the user isnt excluded, has permission and the command isnt in timeout
	if utils.IsExcluded(userId) {
		return nil
	}
	if a.Admin && !utils.IsAdmin(userId) {
		l.Printf("User: %s attempted to execute command %s without authorization", userId, a.Name)
		return ErrNotAuthorized
	} else if a.Admin && utils.IsAdmin(userId) {
		l.Printf("User: %s is executing the admin command %s", userId, a.Name)
	}
	if errT := a.validateTimeout(userId, time.Now().Unix(), l); errT != nil {
		return nil
	}

	//Here we select the type of action and execute it
	switch a.Type {
	case "response":
		errR := a.ResponseAction(userId, chat, author, l)
		if errR != nil {
			return errR
		}
	case "setupgame":
		a.SetUpGameAction(msg)
	default:
		return ErrActionTypeNotFound
	}

	//We register this call to the command to lastCalled globally and by user
	a.LastCalled = time.Now().Unix()
	if a.UserList == nil {
		a.UserList = make(map[string]int64)
	}
	a.UserList[userId] = a.LastCalled
	return nil
}

func (a *Action) ResponseAction(userId string, chat string, author string, l *log.Logger) error {
	r := a.Message
	if strings.Contains(r, "{user}") {
		uname := utils.GetUserName(userId)
		if uname == "" {
			var errU error
			uname, errU = youtubeapi.GetUserFromChannelId(userId, l)
			if errU != nil {
				uname = "virigamer!"
			} else {
				utils.AddToUsers(userId, uname)
			}
		}
		r = strings.ReplaceAll(r, "{user}", uname)
	}
	if strings.Contains(r, "{game}") {
		r = strings.ReplaceAll(r, "{game}", utils.Game)
	}
	err := youtubeapi.PostComment(r, chat, author, l)
	if err != nil {
		return err
	}
	return nil
}

//A positive return value is the number of seconds until the function is available again
//A zero or negative value means the function can be called
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

func (a *Action) SetUpGameAction(msg string) {
	if len(msg) < 8 {
		return
	}
	r := []rune(msg)
	trueName := string(r[7:])
	log.Println(trueName)
	utils.Game = trueName
}
