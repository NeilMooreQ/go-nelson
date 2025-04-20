package pkg

import (
	"encoding/json"
	"os"

	"go-nelson/pkg/structures"
)

var Discord structures.DiscordConfigStruct
var Telegram structures.TelegramConfigStruct
var MongoDB structures.MongoDBConfigStruct
var GoogleAistudio structures.GoogleAistudioConfigStruct
var Parsers structures.ParsersConfigStruct

func LoadConfig(filename string) error {
	if filename == "" {
		filename = "configs.json"
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var config structures.ConfigStruct
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	Discord = config.Discord
	Telegram = config.Telegram
	MongoDB = config.MongoDB
	GoogleAistudio = config.GoogleAistudio
	Parsers = config.Parsers

	return nil
}
