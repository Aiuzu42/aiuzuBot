package bot

import (
	"log"
	"strings"
	"time"

	"github.com/aiuzu42/aiuzuBot/bot/utils"
	"github.com/aiuzu42/aiuzuBot/bot/youtubeapi"
)

type Bot struct {
	//The id of the account used by the bot.
	author string

	//The Id of the chat the bot is in.
	chatId string

	//A pointer to a logger
	logTo *log.Logger

	//Boolean variable to deactive the main loop
	deactivate bool

	//A variable that indicates that the loop function is alredy running
	looping bool

	//A string slice containing a list of admin users id
	admins []string

	//A slice of actions configured for the bot
	actions []Action

	//A timer for the timed actions
	timer int64

	//Token for oauth youtube connection
	token string

	//ApiKey for youtube connection
	apiKey string

	//Current game being played
	game string

	//A slice of quotes for timed actions
	quotes []string

	//List of users excluded by the bot responses
	excluded []string

	//Message filters configurations
	filters utils.Filters

	//Regular expressions used by the bot filters
	matcher []utils.Matcher

	timed []TimedAction

	onFirstMessages bool
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
	bot.logTo = log
	err = bot.refreshToken(config.Configuration.ClientId, config.Configuration.ClientS, config.Configuration.Refresh)
	if err != nil {
		log.Println("Cant initiate bot since we are unable to get a new token")
		return Bot{}, err
	}
	bot.chatId = chatId
	bot.author = config.Configuration.AuthorId
	bot.deactivate = false
	bot.looping = false
	bot.admins = config.Configuration.Admins
	bot.timer = 0
	bot.apiKey = config.Configuration.ApiKey
	bot.quotes = config.Quotes
	bot.excluded = config.Configuration.Excluded
	bot.excluded = append(bot.excluded, bot.author)
	bot.filters = config.Filter
	bot.onFirstMessages = false
	for _, a := range config.Actions {
		bot.actions = append(bot.actions,
			Action{Name: a.Name, Keywords: a.Keywords, Type: a.Type, Message: a.Message, UserTimeout: a.UserTimeout, GlobalTimeout: a.GlobalTimeout, Admin: a.Admin, Uses: a.Uses})
	}
	for _, b := range bot.filters.Word.BanList {
		bot.matcher = append(bot.matcher, utils.NewMatcher(b.Words, bot.logTo))
	}
	for _, t := range config.Timed {
		bot.timed = append(bot.timed, TimedAction{Name: t.Name, Type: t.Type, Cooldown: t.Cooldown, Messages: t.Messages})
	}
	return bot, nil
}

//Loop is the main function of the bot, it reads and process comments and handle events.
//If the function has alredy been called and is looping an error will be returned.
func (b *Bot) Loop() {
	if b.looping {
		b.logTo.Println("Alredy looping.")
		return
	}

	b.executeTimed("first")

	b.looping = true
	b.deactivate = false
	b.onFirstMessages = true

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
				if !b.filter(mi) {
					continue
				}
				for i := range b.actions {
					if b.actions[i].findKeyword(mi.Snippet.DisplayMessage) {
						errA := b.executeAction(mi.Snippet.Author, &b.actions[i])
						if errA != nil {
							b.logTo.Println("Error executing action")
						}
					}
				}
				if b.onFirstMessages {
					b.executeTimed("onFirstMessages")
					b.onFirstMessages = false
				}
			}
		}
		b.executeTimed("timed")
		time.Sleep(10 * time.Second)
	}
	b.deactivate = true
	b.looping = false
	b.executeTimed("ending")
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

func (b *Bot) postTimedAction() {
	msg := utils.GetRandomElement(b.quotes)
	err := b.responseFunction("", msg)
	if err != nil {
		b.logTo.Println("Error posting timed action")
	}
}

//refreshToken obtains a new oauth token.
//All the parameters are optional, if any parameter is an empty string the configuration
//will be loaded and the parameters will be obtained from there.
func (b *Bot) refreshToken(clientId string, secret string, refresh string) error {
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
	token, err := youtubeapi.GetNewAuthToken(clientId, secret, refresh, b.logTo)
	if err != nil {
		return err
	}
	b.token = token
	return nil
}

//executeAction executes the action passed as parameter.
//Validations are made to ensure that:
//-The userId is not in the excluded list.
//-The userId is admin (if it aplies).
//-The action is not in timeout.
//After the action is executed the timeouts are updated.
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
	if !a.validateUses() {
		b.logTo.Println("An action was attemted but it had no more uses")
		return nil
	}
	if errT := a.validateTimeout(userId, time.Now().Unix(), b.logTo); errT != nil {
		return nil
	}

	switch a.Type {
	case "response":
		errR := b.responseFunction(userId, a.Message)
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

//responseFunction is a wrapper function to the PostMessage functionality.
//It takes as input parameters a userId and a message.
//This method replaces the bot variables {user} {game} if present with its correspondent values.
//If the message contains the variable {user} it looks it up in the users table, if its not
//found it retrieves it with the youtubeapi and updates the table.
//In case PostComment fails due to authorization issues, an attempt is made to refresh the
//token and if its successful, a second attempt is made to PostComment.
func (b *Bot) responseFunction(userId string, r string) error {
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
	if err != nil && err == youtubeapi.ErrUnauthorized {
		errR := b.refreshToken("", "", "")
		if errR != nil {
			b.logTo.Println("Cant refresh token")
			return err
		} else {
			b.logTo.Println("Token refresh succesful")
			youtubeapi.PostComment(r, b.chatId, b.author, b.apiKey, b.token, b.logTo)
		}
	} else if err != nil {
		return err
	}
	return nil
}

//deleteFunction is a wrapper function to the DeleteCommment functionality.
//It takes as input parameters a messageId to delete.
//In case DeleteCommment fails due to authorization issues, an attempt is made to refresh the
//token and if its successful, a second attempt is made to DeleteCommment.
func (b *Bot) deleteFunction(msgId string) error {
	err := youtubeapi.DeleteCommment(msgId, b.apiKey, b.token, b.logTo)
	if err != nil && err == youtubeapi.ErrUnauthorized {
		errR := b.refreshToken("", "", "")
		if errR != nil {
			b.logTo.Println("Cant refresh token")
			return err
		} else {
			b.logTo.Println("Token refresh succesful")
			youtubeapi.DeleteCommment(msgId, b.apiKey, b.token, b.logTo)
		}
	} else if err != nil {
		return err
	}
	return nil
}

//penaltyFunction is a wrapper function to the BanUser functionality.
//It takes as input parameters a userId to ban, the type t of ban, and a duration d.
//The banId returned by the api is logged to the the bot logger.
//If the type is youtubeapi.permanent_ban the duration is not used.
//In case BanUser fails due to authorization issues, an attempt is made to refresh the
//token and if its successful, a second attempt is made to BanUser.
func (b *Bot) penaltyFunction(userId string, t string, d int) error {
	banId, err := youtubeapi.BanUser(b.chatId, t, userId, d, b.apiKey, b.token, b.logTo)
	if err != nil && err == youtubeapi.ErrUnauthorized {
		errR := b.refreshToken("", "", "")
		if errR != nil {
			b.logTo.Println("Cant refresh token")
			return err
		} else {
			b.logTo.Println("Token refresh succesful")
			banId, _ = youtubeapi.BanUser(b.chatId, t, userId, d, b.apiKey, b.token, b.logTo)
		}
	} else if err != nil {
		b.logTo.Println("Cant ban user " + userId)
		return err
	}
	b.logTo.Println("User " + userId + " was succesfully banned with banId " + banId)
	return nil
}

func (b *Bot) filter(msg youtubeapi.MessageItem) bool {
	res := true
	if utils.ExistsInSlice(msg.Snippet.Author, b.admins) {
		return res
	}
	if b.filters.Caps.Active {
		res = utils.ValidateCaps(b.filters.Caps.Percent, b.filters.Caps.Min, msg.Snippet.DisplayMessage)
		if !res {
			b.logTo.Printf("Message [%s] didnt pass caps validation", msg.Snippet.DisplayMessage)
			b.deleteFunction(msg.Id)
			if b.filters.Caps.Penalty.Type != "" {
				b.logTo.Printf("A caps penalty was applied for message [%s]", msg.Snippet.DisplayMessage)
				b.penaltyFunction(msg.Snippet.Author, b.filters.Caps.Penalty.Type, b.filters.Caps.Penalty.Duration)
			}
			b.logTo.Printf("A response was send for message [%s]", msg.Snippet.DisplayMessage)
			b.responseFunction(msg.Snippet.Author, b.filters.Caps.Message)
			return res
		}
	}
	if b.filters.Word.Active {
		for i, w := range b.filters.Word.BanList {
			found := b.matcher[i].Match(msg.Snippet.DisplayMessage)
			if found {
				b.logTo.Printf("Message [%s] didnt pass words validation", msg.Snippet.DisplayMessage)
				b.deleteFunction(msg.Id)
				if w.Penalty.Type != "" {
					b.logTo.Printf("A words penalty was applied for message [%s]", msg.Snippet.DisplayMessage)
					b.penaltyFunction(msg.Snippet.Author, w.Penalty.Type, w.Penalty.Duration)
				}
				b.logTo.Printf("A response was send for message [%s]", msg.Snippet.DisplayMessage)
				b.responseFunction(msg.Snippet.Author, w.Message)
				return false
			}
		}
	}
	if b.filters.Max.Active {
		if len(msg.Snippet.DisplayMessage) >= b.filters.Max.Max {
			b.logTo.Printf("Message [%s] didnt pass length validation", msg.Snippet.DisplayMessage)
			b.deleteFunction(msg.Id)
			if b.filters.Max.Penalty.Type != "" {
				b.logTo.Printf("A length penalty was applied for message [%s]", msg.Snippet.DisplayMessage)
				b.penaltyFunction(msg.Snippet.Author, b.filters.Max.Penalty.Type, b.filters.Max.Penalty.Duration)
			}
			b.logTo.Printf("A response was send for message [%s]", msg.Snippet.DisplayMessage)
			b.responseFunction(msg.Snippet.Author, b.filters.Max.Message)
			return false
		}
	}
	return res
}

func (b *Bot) executeTimed(t string) {
	now := time.Now().Unix()
	for i := range b.timed {
		if b.timed[i].Type == t && b.timed[i].Type == "timed" {
			rem := remainingTimeout(now, b.timed[i].Cooldown, b.timed[i].LastCalled)
			if rem <= 0 {
				b.responseFunction("", utils.GetRandomElement(b.timed[i].Messages))
				b.timed[i].LastCalled = now
			}
		} else if b.timed[i].Type == t {
			b.responseFunction("", utils.GetRandomElement(b.timed[i].Messages))
			b.timed[i].LastCalled = now
		}
	}
}
