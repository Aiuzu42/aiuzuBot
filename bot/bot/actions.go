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
	if utils.IsExcluded(userId) {
		return nil
	}
	if a.Admin && !utils.IsAdmin(userId) {
		l.Printf("User: %s attempted to execute command %s without authorization", userId, a.Name)
		return ErrNotAuthorized //TODO post comment about it? or just dont do it?
	} else if a.Admin && utils.IsAdmin(userId) {
		l.Printf("User: %s executed the admin command %s", userId, a.Name)
	}
	if errT := a.validateTimeout(userId, time.Now().Unix()); errT != nil {
		//youtubeapi.PostComment(errT.Error(), chat, author, l)
		l.Println(errT.Error())
		return nil
	}
	switch a.Type {
	case "response":
		a.ResponseAction(userId, chat, author, l)
	case "setupgame":
		a.SetUpGameAction(msg)
		return nil
	default:
		return ErrActionTypeNotFound
	}
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
				uname = "Virigamer!"
			} else {
				utils.AddToUsers(userId, uname)
			}
		}
		r = strings.ReplaceAll(r, "{user}", uname)
	} else if strings.Contains(r, "{game}") {
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

func (a *Action) validateTimeout(user string, now int64) error {
	global := remainingTimeout(now, a.GlobalTimeout, a.LastCalled)
	if global > 0 {
		e := fmt.Sprintf("%d seconds remaining to user %s command again[global].", global, a.Name)
		return errors.New(e)
	}
	ut := a.UserList[user]
	timeUser := remainingTimeout(now, a.UserTimeout, ut)
	if timeUser > 0 {
		//TODO personalizar por nombre de usuario
		e := fmt.Sprintf("%d seconds remaining to user %s command again[user].", global, a.Name)
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
