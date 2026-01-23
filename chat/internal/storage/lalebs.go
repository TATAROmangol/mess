package storage

const (
	AllLabelsSelect = "*"
	ReturningSuffix = "RETURNING *"
	SkipLocked      = "FOR UPDATE SKIP LOCKED"
	IsNullLabel     = "IS NULL"
	AscSortLabel    = "ASC"
	DescSortLabel   = "DESC"
)

type Table = string

const (
	ChatTable           Table = "chat"
	LastReadTable       Table = "last_read"
	MessageTable        Table = "message"
	MessageOutboxTable  Table = "message_outbox"
	LastReadOutboxTable Table = "last_read_outbox"
)

type Label = string

// ChatTable
const (
	ChatIDLabel              Label = "id"
	ChatFirstSubjectIDLabel  Label = "first_subject_id"
	ChatSecondSubjectIDLabel Label = "second_subject_id"
	ChatMessagesCount        Label = "messages_count"
	ChatUpdatedAtLabel       Label = "updated_at"
	ChatCreatedAtLabel       Label = "created_at"
	ChatDeletedAtLabel       Label = "deleted_at"
)

// LastReadTable
const (
	LastReadSubjectIDLabel     Label = "subject_id"
	LastReadChatIDLabel        Label = "chat_id"
	LastReadMessageIDLabel     Label = "message_id"
	LastReadMessageNumberLabel Label = "message_number"
	LastReadUpdatedAtLabel     Label = "updated_at"
	LastReadDeletedAtLabel     Label = "deleted_at"
)

// MessageTable
const (
	MessageIDLabel              Label = "id"
	MessageChatIDLabel          Label = "chat_id"
	MessageSenderSubjectIDLabel Label = "sender_subject_id"
	MessageContentLabel         Label = "content"
	MessageNumberLabel          Label = "number"
	MessageVersionLabel         Label = "version"
	MessageCreatedAtLabel       Label = "created_at"
	MessageUpdatedAtLabel       Label = "updated_at"
	MessageDeletedAtLabel       Label = "deleted_at"
)

// MessageOutboxTable
const (
	MessageOutboxIDLabel          Label = "id"
	MessageOutboxRecipientIDLabel Label = "recipient_id"
	MessageOutboxMessageIDLabel   Label = "message_id"
	MessageOutboxOperationLabel   Label = "operation"
	MessageOutboxDeletedAtLabel   Label = "deleted_at"
)

// LastReadOutboxTable
const (
	LastReadOutboxIDLabel          Label = "id"
	LastReadOutboxRecipientIDLabel Label = "recipient_id"
	LastReadOutboxChatIDLabel      Label = "chat_id"
	LastReadOutboxSubjectIDLabel   Label = "subject_id"
	LastReadOutboxMessageIDLabel   Label = "message_id"
	LastReadOutboxDeletedAtLabel   Label = "deleted_at"
)
