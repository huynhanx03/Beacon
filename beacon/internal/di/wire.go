package di

import (
	"beacon/config"
	"beacon/internal/adapters/driven/notifier"
	"beacon/internal/adapters/driven/provider"
	"beacon/internal/core/service"
)

func Wire(cfg *config.Config) *service.BeaconService {
	discord := notifier.NewDiscordNotifier(cfg.DiscordWebhookURL)
	leetcode := provider.NewLeetCodeProvider()
	return service.NewBeaconService(discord, leetcode, cfg.Reminder)
}
