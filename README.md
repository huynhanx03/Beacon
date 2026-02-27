# Beacon

A lightweight reminder and notification assistant, built to keep you on track every day.

## Features

### Reminders
- [x] **DSA** — Daily LeetCode challenge with difficulty, tags, and direct link
- [x] **Healthy** — Randomized health tips (hydration, posture, eye rest, stretching, breathing)
- [x] **Review** — Nightly self-reflection prompts to keep you accountable

## Tech Stack

- **Language:** Go 1.25+
- **Scheduling:** GitHub Actions (cron)
- **Notification:** Discord Webhooks

## Setup

### 1. Repository Secrets

| Secret | Description |
|---|---|
| `BEACON_DSA_WEBHOOK_URL` | Discord webhook for DSA reminders |
| `BEACON_HEALTHY_WEBHOOK_URL` | Discord webhook for health reminders |
| `BEACON_REVIEW_WEBHOOK_URL` | Discord webhook for review reminders |

### 2. Run Locally

```bash
cd beacon
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."
export REMINDER_TYPE="dsa"  # dsa | healthy | review
go run ./cmd/beacon
```

## License

[MIT](LICENSE)