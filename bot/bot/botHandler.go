package bot

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

const (
	prefix = "botconfig-"
	suffix = ".json"
)

type BotHandler struct {
	bots     []Bot
	settings GlobalConfig
	logTo    *log.Logger
}

func NewBotHandler(log *log.Logger) BotHandler {
	bh := BotHandler{logTo: log, settings: GlobalConfig{}}
	file, err := os.Open("./")
	if err != nil {
		bh.logTo.Println("Error opening directory: " + err.Error())
	} else {
		defer file.Close()
		names, err := file.Readdirnames(0)
		if err != nil {
			bh.logTo.Println("Error reading file names: " + err.Error())
		} else {
			for _, n := range names {
				if strings.HasPrefix(n, prefix) && strings.HasSuffix(n, suffix) {
					idFromFile := stripIDFromFile(n)
					lc, errR := loadLocalConfig(idFromFile, bh.logTo)
					if errR != nil {
						continue
					}
					if idFromFile != lc.BotId {
						bh.logTo.Println("Id in filename and config does not coincide in " + n)
						continue
					}
					if bh.doesBotExists(idFromFile) {
						bh.logTo.Println("A bot with that name alredy exists: " + idFromFile)
						continue
					}
					bh.settings.Global = append(bh.settings.Global, SimpleBotId{BotId: idFromFile, BotType: getBotType(lc.Type)})
				}
			}
		}
	}
	return bh
}

func (bh *BotHandler) doesBotExists(name string) bool {
	for _, i := range bh.settings.Global {
		if i.BotId == name {
			return true
		}
	}
	return false
}

func (bh *BotHandler) updateGame(botId string, game string) error {
	for i := range bh.bots {
		if bh.bots[i].BotId == botId {
			bh.bots[i].game = game
			return nil
		}
	}
	return ErrorFindingBot
}

func (bh *BotHandler) getGame(botId string) (string, error) {
	for i := range bh.bots {
		if bh.bots[i].BotId == botId {
			return bh.bots[i].game, nil
		}
	}
	return "", ErrorFindingBot
}

func (bh *BotHandler) startBot(botId string, liveId string, game string) error {
	bh.logTo.Println("We just enter startBot")
	if !bh.doesBotExists(botId) {
		bh.logTo.Println("The bot you want to start does not exists: " + botId)
		return ErrorFindingBot
	}
	for i := range bh.bots {
		if bh.bots[i].BotId == botId {
			bh.logTo.Println("Th bot is alredy looping: " + botId)
			return ErrorBotAlredyExists
		}
	}
	lc, err := loadLocalConfig(botId, bh.logTo)
	if err != nil {
		return err
	}
	bot, err := NewBot(lc, liveId, bh.logTo)
	if err != nil {
		return err
	}
	bh.bots = append(bh.bots, bot)
	go bh.bots[len(bh.bots)-1].Loop()
	bh.logTo.Println("We are about to exit startBot")
	if game != "" {
		bh.updateGame(botId, game)
	}
	return nil
}

func (bh *BotHandler) stopBot(botId string) error {
	bh.logTo.Println("We just enter stopBot")
	for i := range bh.bots {
		if bh.bots[i].BotId == botId {
			bh.bots[i].DeactivateLoop()
			copy(bh.bots[i:], bh.bots[i+1:])
			bh.bots = bh.bots[:len(bh.bots)-1]
			bh.logTo.Println("We are about to exit stopBot")
			return nil
		}
	}
	return ErrorFindingBot
}

func (bh *BotHandler) getSimpleBotList() GlobalConfig {
	return bh.settings
}

func (bh *BotHandler) getBotConfiguration(botId string) (LocalConfig, error) {
	return loadLocalConfig(botId, bh.logTo)
}

func (bh *BotHandler) saveNewConfiguration(lc LocalConfig) error {
	if bh.doesBotExists(lc.BotId) {
		bh.logTo.Printf("Bot with name %s alredy exists.", lc.BotId)
		return ErrorBotAlredyExists
	}
	err := bh.validateAndSaveConfiguration(lc)
	if err != nil {
		return err
	}
	bh.settings.Global = append(bh.settings.Global, SimpleBotId{BotId: lc.BotId, BotType: getBotType(lc.Type)})
	return nil
}

//You cant change the name of a bot
func (bh *BotHandler) updateConfiguration(lc LocalConfig) error {
	if !bh.doesBotExists(lc.BotId) {
		bh.logTo.Printf("Bot with name %s does not exists, cant update configuration.", lc.BotId)
		return ErrorFindingBot
	}
	err := bh.validateAndSaveConfiguration(lc)
	if err != nil {
		return err
	}
	for i := range bh.settings.Global {
		if bh.settings.Global[i].BotId == lc.BotId {
			bh.settings.Global[i].BotType = lc.Type
			break
		}
	}
	return nil
}

func (bh *BotHandler) validateAndSaveConfiguration(lc LocalConfig) error {
	bh.logTo.Printf("Validating bot configuration: [%s]", lc.BotId)
	if lc.validate(bh.logTo) {
		bh.logTo.Println(ErrorValidating.Error())
		return ErrorValidating
	}
	err := bh.writeDataToFile(lc.BotId)
	if err != nil {
		bh.logTo.Println(ErrorSavingConfig.Error())
		return ErrorSavingConfig
	}
	bh.logTo.Printf("Bot configuration succesfully saved: [%s]", lc.BotId)
	return nil
}

func (bh *BotHandler) writeDataToFile(name string) error {
	file, _ := os.OpenFile(prefix+name+suffix, os.O_CREATE, os.ModePerm)
	defer file.Close()
	err := json.NewEncoder(file).Encode(bh.settings)
	if err != nil {
		bh.logTo.Println("Unable to save data.")
		return err
	}
	return nil
}

func stripIDFromFile(fileName string) string {
	return fileName[10 : len(fileName)-5]
}

func getBotType(botType string) string {
	if botType == "twitch" || botType == "youtube" {
		return botType
	} else {
		return "undefined"
	}
}
