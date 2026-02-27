package config

import "os"

type ReminderType string

const (
	ReminderDSA     ReminderType = "dsa"
	ReminderHealthy ReminderType = "healthy"
	ReminderReview  ReminderType = "review"
)

type Config struct {
	DiscordWebhookURL string
	Reminder          ReminderType
}

func Load() *Config {
	return &Config{
		DiscordWebhookURL: getEnv("DISCORD_WEBHOOK_URL", ""),
		Reminder:          ReminderType(getEnv("REMINDER_TYPE", "dsa")),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
