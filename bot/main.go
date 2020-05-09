package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aiuzu42/aiuzuBot/bot/bot"
	"github.com/aiuzu42/aiuzuBot/bot/utils"
)

var Bot bot.Bot

func main() {
	//Prepare log
	now := time.Now()
	fileName := "commentsLog" + now.Format("020120061504") + ".txt"
	f, errF := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if errF != nil {
		log.Fatal("Error reading log file.")
	}
	defer f.Close()
	l := log.New(f, "[aiuzuBot] ", log.LstdFlags)

	//Setup bot and start loop
	Bot = setupBot(l)
	go Bot.Loop()

	//Setup web server for configuration
	http.HandleFunc("/", handlerFunc)
	http.ListenAndServe(":8080", nil)
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", 404)
		return
	}
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "templates/index.html")
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "There was an error parsing the request form: %v", err)
			return
		}
		Bot.UpdateGame(r.FormValue("iGame"))
		http.ServeFile(w, r, "templates/index.html")
	default:
		http.Error(w, "405 Method Not Allowed", 405)
		return
	}
}

func setupBot(l *log.Logger) bot.Bot {
	config, err := utils.LoadConfig()
	if err != nil {
		l.Fatal("Cant read configuration")
	}
	bot, err := bot.NewBot(config, l)
	if err != nil {
		l.Fatal(err.Error())
	}
	return bot
}
