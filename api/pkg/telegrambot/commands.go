package telegrambot

import (
	"fmt"
	"strings"
)

// BotCommand is one Telegram slash command for setMyCommands.
type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// DefaultBotCommands returns the bot's slash commands for Telegram registration.
func DefaultBotCommands() []BotCommand {
	return []BotCommand{
		{Command: "price", Description: "Cheapest in-stock match"},
		{Command: "help", Description: "Show available commands"},
	}
}

func formatHelpMessage() string {
	lines := []string{
		"Search Singapore MTG singles prices.",
		"",
		"Commands:",
	}
	for _, cmd := range DefaultBotCommands() {
		lines = append(lines, fmt.Sprintf("/%s — %s", cmd.Command, cmd.Description))
	}
	lines = append(lines, "", "Example: /price Lightning Bolt")
	return strings.Join(lines, "\n")
}
