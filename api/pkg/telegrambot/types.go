package telegrambot

// Update is a subset of the Telegram webhook update payload.
type Update struct {
	Message *Message `json:"message"`
}

// Message is a subset of a Telegram message.
type Message struct {
	MessageID      int64    `json:"message_id"`
	Chat           Chat     `json:"chat"`
	Text           string   `json:"text"`
	ReplyToMessage *Message `json:"reply_to_message"`
}

// Chat identifies a Telegram chat.
type Chat struct {
	ID int64 `json:"id"`
}
