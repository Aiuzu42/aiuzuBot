package bot

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type responseError struct {
	Message string `json:"message"`
}

func (bh *BotHandler) SimpleBotListEndpoint(w http.ResponseWriter, r *http.Request) {
	gc := bh.getSimpleBotList()
	if len(gc.Global) < 1 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		json.NewEncoder(w).Encode(gc.Global)
	}
}

func (bh *BotHandler) StartBotEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	liveId := r.URL.Query().Get("liveId")
	game := r.URL.Query().Get("game")
	err := bh.startBot(params["botid"], liveId, game)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseError{Message: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (bh *BotHandler) StopBotEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	err := bh.stopBot(params["botid"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseError{Message: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (bh *BotHandler) GetBotConfigEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	lc, err := bh.getBotConfiguration(params["botid"])
	if err != nil {
		if err == UnableToLoadConfig {
			w.WriteHeader(http.StatusNotFound)
		} else if err == UnableToDecodeConfig {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(responseError{Message: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lc)
}

func (bh *BotHandler) GetBotInfoEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	game, err := bh.getGame(params["botid"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(err.Error())
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(game))
}

func (bh *BotHandler) UpdateBotConfigEndpoint(w http.ResponseWriter, r *http.Request) {
	var lc LocalConfig
	err := json.NewDecoder(r.Body).Decode(&lc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responseError{Message: UnableToDecodeConfig.Error()})
		return
	}
	err = bh.updateConfiguration(lc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responseError{Message: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (bh *BotHandler) UpdateBotInfoEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	game := r.URL.Query().Get("game")
	if game == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Query param game is mandatory."))
		return
	}
	err := bh.updateGame(params["botid"], game)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(responseError{Message: err.Error()})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (bh *BotHandler) AddNewBotEndpoint(w http.ResponseWriter, r *http.Request) {
	var lc LocalConfig
	err := json.NewDecoder(r.Body).Decode(&lc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responseError{Message: UnableToDecodeConfig.Error()})
		return
	}
	err = bh.saveNewConfiguration(lc)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(responseError{Message: err.Error()})
		return
	}
	w.WriteHeader(http.StatusCreated)
}
