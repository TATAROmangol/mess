package httpdto

type ProfileResponse struct {
	SubjectID string `json:"subject_id"`
	Alias     string `json:"alias"`
	AvatarURL string `json:"avatar_url"`
	Version   int    `json:"version"`
}

type ProfilesResponse struct {
	Profiles []*ProfileResponse `json:"profiles"`
}

type AddProfileRequest struct {
	Alias string `json:"alias"`
}

type UpdateProfileMetadataRequest struct {
	Alias   string `json:"alias"`
	Version int    `json:"version"`
}

type UploadAvatarResponse struct {
	UploadURL string `json:"upload_url"`
}