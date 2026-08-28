package telegrambot

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultBotCommands(t *testing.T) {
	commands := DefaultBotCommands()
	require.Len(t, commands, 3)
	require.Equal(t, "price", commands[0].Command)
	require.Equal(t, "ck", commands[1].Command)
	require.Equal(t, "help", commands[2].Command)
}

func TestTelegramMenuCommands_OmitsRequiredArgCommands(t *testing.T) {
	commands := TelegramMenuCommands()
	require.Len(t, commands, 1)
	require.Equal(t, "help", commands[0].Command)
}

func TestFormatHelpMessage_IncludesCommands(t *testing.T) {
	msg := formatHelpMessage()
	for _, cmd := range DefaultBotCommands() {
		require.Contains(t, msg, "/"+cmd.Command)
		require.Contains(t, msg, cmd.Description)
	}
	require.Contains(t, msg, "/price Lightning Bolt")
	require.Contains(t, msg, "/ck Lightning Bolt")
	require.True(t, strings.HasPrefix(msg, "Search Singapore MTG singles prices."))
}
