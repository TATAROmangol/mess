package storage

const (
	AllLabelsSelect = "*"
	ReturningSuffix = "RETURNING *"
	IsNullLabel     = "IS NULL"
)

type Table = string

const (
	ProfileTable         Table = "profile"
	AvatarKeyOutboxTable Table = "avatar_key_outbox"
)

type Label = string

// Profile
const (
	ProfileSubjectIDLabel Label = "subject_id"
	ProfileAliasLabel     Label = "alias"
	ProfileAvatarKeyLabel Label = "avatar_key"
	ProfileVersionLabel   Label = "version"
	ProfileUpdatedAtLabel Label = "updated_at"
	ProfileCreatedAtLabel Label = "created_at"
	ProfileDeletedAtLabel Label = "deleted_at"
)

// AvatarKeyOutbox
const (
	AvatarKeyOutboxKeyLabel       Label = "key"
	AvatarKeyOutboxSubjectIDLabel Label = "subject_id"
	AvatarKeyOutboxDeletedAtLabel Label = "deleted_at"
	AvatarKeyOutboxCreatedAtLabel Label = "created_at"
)
