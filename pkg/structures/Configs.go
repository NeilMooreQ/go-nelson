package structures

type DiscordConfigStruct struct {
	Token       string `json:"token"`
	GuildID     string `json:"guild_id"`
	Enabled     bool   `json:"enabled"`
	NewsForumId string `json:"news_forum_id"`
}

type TelegramConfigStruct struct {
	Token     string `json:"token"`
	Enabled   bool   `json:"enabled"`
	ChannelID string `json:"channel_id"`
}

type MongoDBConfigStruct struct {
	URI        string `json:"uri"`
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

type GoogleAistudioConfigStruct struct {
	APIKey string `json:"api_key"`
}

type ParsersConfigStruct struct {
	ThreeDNews      bool `json:"3dnews"`
	DisgustingMen   bool `json:"disgustingmen"`
	DTF             bool `json:"dtf"`
	EpicGames       bool `json:"epicgames"`
	GamedevRu       bool `json:"gamedevru"`
	Ixbt            bool `json:"ixbt"`
	SteamDevelopers bool `json:"steam_developers"`
	Stopgame        bool `json:"stopgame"`
}

type ConfigStruct struct {
	Discord        DiscordConfigStruct        `json:"discord"`
	Telegram       TelegramConfigStruct       `json:"telegram"`
	MongoDB        MongoDBConfigStruct        `json:"mongodb"`
	GoogleAistudio GoogleAistudioConfigStruct `json:"google_aistudio"`
	Parsers        ParsersConfigStruct        `json:"parsers"`
}
