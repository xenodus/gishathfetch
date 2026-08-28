package telegrambot

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultBotCommands(t *testing.T) {
	commands := DefaultBotCommands()
	require.Len(t, commands, 2)
	require.Equal(t, "price", commands[0].Command)
	require.Equal(t, "help", commands[1].Command)
}

func TestFormatHelpMessage_IncludesCommands(t *testing.T) {
	msg := formatHelpMessage()
	for _, cmd := range DefaultBotCommands() {
		require.Contains(t, msg, "/"+cmd.Command)
		require.Contains(t, msg, cmd.Description)
	}
	require.Contains(t, msg, "Example: /price Lightning Bolt")
	require.True(t, strings.HasPrefix(msg, "Search Singapore MTG singles prices."))
}
