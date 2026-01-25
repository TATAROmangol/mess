package s3client

type UploadEventMinIO struct {
	Records []struct {
		S3 struct {
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

func (uemio *UploadEventMinIO) GetKey() string {
	return uemio.Records[0].S3.Object.Key
}
