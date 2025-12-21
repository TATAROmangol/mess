package profile

type Label = string

const (
	SubjectIDLabel Label = "subject_id"
	AliasLabel     Label = "alias"
	AvatarURLLabel Label = "avatar_url"
	VersionLabel   Label = "version"
	UpdatedAtLabel Label = "updated_at"
	CreatedAtLabel Label = "created_at"
)

type Table = string

const (
	ProfileTable Table = "profile"
)
