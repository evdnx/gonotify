package messenger

// Messenger defines the contract for sending messages to different platforms.
type Messenger interface {
	SendMessage(message string) error
	Name() string
}

