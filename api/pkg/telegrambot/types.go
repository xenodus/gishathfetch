package telegrambot

// Update is a subset of the Telegram webhook update payload.
type Update struct {
	Message *Message `json:"message"`
}

// Message is a subset of a Telegram message.
type Message struct {
	MessageID      int64    `json:"message_id"`
	From           *User    `json:"from"`
	Chat           Chat     `json:"chat"`
	Text           string   `json:"text"`
	ReplyToMessage *Message `json:"reply_to_message"`
}

// User identifies a Telegram user.
type User struct {
	ID int64 `json:"id"`
}

// SenderID returns the Telegram user ID for the message sender.
func (m *Message) SenderID() int64 {
	if m == nil || m.From == nil {
		return 0
	}
	return m.From.ID
}

// Chat identifies a Telegram chat.
type Chat struct {
	ID int64 `json:"id"`
}
