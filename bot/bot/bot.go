package bot

import (
	"log"
	"time"

	"github.com/aiuzu42/aiuzuBot/bot/utils"
	"github.com/aiuzu42/aiuzuBot/bot/youtubeapi"
)

type Bot struct {
	Author     string
	ChatId     string
	logTo      *log.Logger
	deactivate bool
	looping    bool
	admins     []string
	actions    []Action
	timer      int64
}

func NewBot(a string, c string, act []Action, l *log.Logger) Bot {
	return Bot{a, c, l, false, false, nil, act, 0}
}

func (b *Bot) InitialHello() error {
	err := youtubeapi.PostComment(utils.InitialHello, b.ChatId, b.Author, b.logTo)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bot) Loop() {
	if b.looping {
		b.logTo.Println("Alredy looping.")
		return
	}
	b.looping = true
	b.deactivate = false
	next := ""
	b.timer = time.Now().Unix()
	for !b.deactivate {
		m, err := youtubeapi.ReadMessages(b.ChatId, next, b.logTo)
		if m.Info.Total > 50 {
			b.logTo.Println("Too many messages, nothing to do this cycle")
			next = m.Next
			time.Sleep(10 * time.Second)
			continue
		}
		if err != nil {
			b.logTo.Println("There was an error attempting to read messages.")
			continue
		}
		next = m.Next
		for _, mi := range m.Messages {
			logMessage(mi, b.logTo)
			for i := range b.actions {
				if b.actions[i].findKeyword(mi.Snippet.DisplayMessage) {
					errA := b.actions[i].ExecuteAction(mi.Snippet.Author, b.ChatId, b.Author, mi.Snippet.DisplayMessage, b.logTo)
					if errA != nil {
						b.logTo.Println(errA.Error())
					}
					continue
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
}

func logMessage(m youtubeapi.MessageItem, l *log.Logger) {
	l.Println("########################################################")
	l.Println("Author: " + m.Snippet.Author)
	l.Println("DisplayMessage: " + m.Snippet.DisplayMessage)
	l.Println("########################################################")
}

func (b *Bot) DeactivateLoop() {
	b.deactivate = true
}

func (b *Bot) postTimedAction() error {
	msg := utils.GetRandomQuote()
	err := youtubeapi.PostComment(msg, b.ChatId, b.Author, b.logTo)
	if err != nil {
		b.logTo.Println("Error posting timed action")
		b.logTo.Println(err.Error())
		return err
	}
	return nil
}
