package structures

import (
	"github.com/qiniu/qmgo/field"
)

type News struct {
	field.DefaultField `bson:",inline"`
	Provider           string   `bson:"provider"`
	UniqueID           string   `bson:"unique_id"`
	Title              string   `bson:"title"`
	Description        string   `bson:"description"`
	URL                string   `bson:"url"`
	Tags               []string `bson:"tags"`
	Images             []string `bson:"images"`
	TelegramMessageID  string   `bson:"telegram_message_id,omitempty"`
	DiscordThreadID    string   `bson:"discord_thread_id,omitempty"`
	DiscordMessageID   string   `bson:"discord_message_id,omitempty"`
}
