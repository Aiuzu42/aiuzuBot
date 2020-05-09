package bot

import (
	"log"
	"strings"
	"time"

	"github.com/aiuzu42/aiuzuBot/bot/utils"
	"github.com/aiuzu42/aiuzuBot/bot/youtubeapi"
)

type Bot struct {
	author       string
	chatId       string
	logTo        *log.Logger
	deactivate   bool
	looping      bool
	admins       []string
	actions      []Action
	timer        int64
	filters      bool
	token        string
	apiKey       string
	game         string
	quotes       []string
	initialHello string
	excluded     []string
}

//NewBot initializes a Bot struct and sets its values based on the configuration and log provided.
//It obtains the liveChatId and a refreshToken from the youtube API.
//If an error ocurrs while obtaining data from the youtube API, an zero value Bot is returned with an error.
func NewBot(config utils.Config, log *log.Logger) (Bot, error) {
	chatId, err := youtubeapi.GetFristLiveChatIdFromChannelId(config.Configuration.LiveStreamChannelId, config.Configuration.ApiKey, log)
	if err != nil {
		log.Println("Cant initiate bot since the channel doesnt have an active livestream")
		return Bot{}, err
	}
	bot := Bot{}
	err = bot.refreshToken(config.Configuration.ClientId, config.Configuration.ClientS, config.Configuration.Refresh, log)
	if err != nil {
		log.Println("Cant initiate bot since we are unable to get a new token")
		return Bot{}, err
	}
	bot.chatId = chatId
	bot.author = config.Configuration.AuthorId
	bot.logTo = log
	bot.deactivate = false
	bot.looping = false
	bot.admins = config.Configuration.Admins
	bot.timer = 0
	bot.filters = false
	bot.apiKey = config.Configuration.ApiKey
	bot.quotes = config.Quotes
	bot.initialHello = config.InitialHello
	bot.excluded = config.Configuration.Excluded
	bot.excluded = append(bot.excluded, bot.author)
	for _, a := range config.Actions {
		bot.actions = append(bot.actions,
			Action{Name: a.Name, Keywords: a.Keywords, Type: a.Type, Message: a.Message, UserTimeout: a.UserTimeout, GlobalTimeout: a.GlobalTimeout, Admin: a.Admin})
	}
	return bot, nil
}

func (b *Bot) sayInitialHello() error {
	err := youtubeapi.PostComment(b.initialHello, b.chatId, b.author, b.apiKey, b.token, b.logTo)
	if err != nil {
		return err
	}
	return nil
}

//Loop is the main function of the bot, it reads and process comments and handle events.
//If the function has alredy been called and is looping an errro will be returned.
func (b *Bot) Loop() {
	if b.looping {
		b.logTo.Println("Alredy looping.")
		return
	}

	errH := b.sayInitialHello()
	if errH != nil {
		b.logTo.Println("Cant say hello")
	}

	b.looping = true
	b.deactivate = false

	next := ""

	b.timer = time.Now().Unix()

	for !b.deactivate {
		m, err := youtubeapi.ReadMessages(b.chatId, next, b.apiKey, b.logTo)
		if err != nil {
			b.logTo.Println("There was an error attempting to read messages.")
		} else if m.Info.Total > 20 {
			b.logTo.Println("Too many messages, nothing to do this cycle")
			next = m.Next
		} else {
			next = m.Next
			for _, mi := range m.Messages {
				logMessage(mi, b.logTo)
				for i := range b.actions {
					if b.actions[i].findKeyword(mi.Snippet.DisplayMessage) {
						errA := b.executeAction(mi.Snippet.Author, &b.actions[i])
						if errA != nil && errA == youtubeapi.ErrUnauthorized {
							errR := b.refreshToken("", "", "", b.logTo)
							if errR != nil {
								b.logTo.Println("Cant refresh token")
							} else {
								b.logTo.Println("Token refresh succesful")
							}
							b.executeAction(mi.Snippet.Author, &b.actions[i])
						} else if errA != nil {
							b.logTo.Println("Error executing action")
						}
					}
				}
			}
		}
		now := time.Now().Unix()
		if now >= b.timer+720 {
			b.timer = now
			b.postTimedAction()
		}
		time.Sleep(10 * time.Second)
	}
	b.deactivate = true
	b.looping = false
}

//DeactivateLoop stops this bot loop.
func (b *Bot) DeactivateLoop() {
	b.deactivate = true
}

//UpdateGame is used to update the name of the current game in the livestream.
func (b *Bot) UpdateGame(g string) {
	b.game = g
}

func logMessage(m youtubeapi.MessageItem, l *log.Logger) {
	l.Println("########################################################")
	l.Println("Author: " + m.Snippet.Author)
	l.Println("DisplayMessage: " + m.Snippet.DisplayMessage)
	l.Println("########################################################")
}

func (b *Bot) postTimedAction() error {
	msg := utils.GetRandomElement(b.quotes)
	err := youtubeapi.PostComment(msg, b.chatId, b.author, b.apiKey, b.token, b.logTo)
	if err != nil && err == youtubeapi.ErrUnauthorized {
		errR := b.refreshToken("", "", "", b.logTo)
		if errR != nil {
			b.logTo.Println("Cant refresh token")
		} else {
			b.logTo.Println("Token refresh succesful")
		}
		youtubeapi.PostComment(msg, b.chatId, b.author, b.apiKey, b.token, b.logTo)
	}
	if err != nil {
		b.logTo.Println("Error posting timed action")
		return err
	}
	return nil
}

func (b *Bot) refreshToken(clientId string, secret string, refresh string, log *log.Logger) error {
	if clientId == "" || secret == "" || refresh == "" {
		config, errC := utils.LoadConfig()
		if errC != nil {
			return errC
		} else {
			clientId = config.Configuration.ClientId
			secret = config.Configuration.ClientS
			refresh = config.Configuration.Refresh
		}

	}
	token, err := youtubeapi.GetNewAuthToken(clientId, secret, refresh, log)
	if err != nil {
		return err
	}
	b.token = token
	return nil
}

func (b *Bot) executeAction(userId string, a *Action) error {
	if utils.ExistsInSlice(userId, b.excluded) {
		return nil
	}
	if a.Admin && !utils.ExistsInSlice(userId, b.admins) {
		b.logTo.Printf("User: %s attempted to execute command %s without authorization", userId, a.Name)
		return ErrNotAuthorized
	} else if a.Admin && utils.ExistsInSlice(userId, b.admins) {
		b.logTo.Printf("User: %s is executing the admin command %s", userId, a.Name)
	}
	if errT := a.validateTimeout(userId, time.Now().Unix(), b.logTo); errT != nil {
		return nil
	}

	switch a.Type {
	case "response":
		errR := b.responseAction(userId, a)
		if errR != nil {
			return errR
		}
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

func (b *Bot) responseAction(userId string, a *Action) error {
	r := a.Message
	if strings.Contains(r, "{user}") {
		uname := utils.GetUserName(userId)
		if uname == "" {
			var errU error
			uname, errU = youtubeapi.GetUserFromChannelId(userId, b.apiKey, b.logTo)
			if errU != nil {
				uname = ""
			} else {
				utils.AddToUsers(userId, uname)
			}
		}
		r = strings.ReplaceAll(r, "{user}", uname)
	}
	if strings.Contains(r, "{game}") {
		r = strings.ReplaceAll(r, "{game}", b.game)
	}
	err := youtubeapi.PostComment(r, b.chatId, b.author, b.apiKey, b.token, b.logTo)
	if err != nil {
		return err
	}
	return nil
}
