package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aiuzu42/aiuzuBot/bot/bot"
	"github.com/gorilla/mux"
)

func main() {
	//Prepare log
	now := time.Now()
	fileName := "mainLog" + now.Format("020120061504") + ".txt"
	f, errF := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if errF != nil {
		log.Fatal("Error reading log file.")
	}
	defer f.Close()

	l := log.New(f, "[aiuzuBot] ", log.LstdFlags)

	bh := bot.NewBotHandler(l)

	router := mux.NewRouter()
	router.Use(routerMiddleware)

	router.HandleFunc("/aiuzubot/v3/list", bh.SimpleBotListEndpoint).Methods("GET")
	router.HandleFunc("/aiuzubot/v3/bot/{botid}/start", bh.StartBotEndpoint).Methods("GET")
	router.HandleFunc("/aiuzubot/v3/bot/{botid}/stop", bh.StopBotEndpoint).Methods("GET")
	router.HandleFunc("/aiuzubot/v3/bot/{botid}/config", bh.GetBotConfigEndpoint).Methods("GET")
	router.HandleFunc("/aiuzubot/v3/bot/{botid}/config", bh.UpdateBotConfigEndpoint).Methods("PUT")
	router.HandleFunc("/aiuzubot/v3/bot/{botid}/info", bh.UpdateBotInfoEndpoint).Methods("PUT")
	router.HandleFunc("/aiuzubot/v3/bot/{botid}/info", bh.GetBotInfoEndpoint).Methods("GET")
	router.HandleFunc("/aiuzubit/v3/bot", bh.AddNewBotEndpoint).Methods("POST")

	http.ListenAndServe(":3000", router)

}

func routerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
