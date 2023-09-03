package slack

type SlackConfig struct {
	SlackChannels []string `envconfig:"SLACK_CHANNELS"`
	BotToken      string   `envconfig:"SLACK_BOT_TOKEN"`
	AppToken      string   `envconfig:"SLACK_APP_TOKEN"`
}
