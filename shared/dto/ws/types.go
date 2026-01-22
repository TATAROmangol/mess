package ws

type Operation string

const (
	UnknownOperation Operation = "unknown"
	SendMessage      Operation = "send_message"
	UpdateOperation  Operation = "update_message"
	UpdateLastRead   Operation = "update_last_read"
)
