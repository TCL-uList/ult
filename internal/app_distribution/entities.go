package appdistribution

type ReleaseNotes struct {
	Text string `json:"text"`
}

type Release struct {
	Name               string       `json:"name"`
	ReleaseNotes       ReleaseNotes `json:"releaseNotes"`
	DisplayVersion     string       `json:"displayVersion"`
	BuildVersion       string       `json:"buildVersion"`
	CreateTime         string       `json:"createTime"`
	FirebaseConsoleUri string       `json:"firebaseConsoleUri"`
	TestingUri         string       `json:"testingUri"`
	BinaryDownloadUri  string       `json:"binaryDownloadUri"`
}

type UploadReleaseResult struct {
	Result  string  `json:"result"`
	Release Release `json:"release"`
}

type OperationResult struct {
	Name string `json:"name"`
	Done bool   `json:"done"`
	// Error   any                 `json:"error"`
	Release UploadReleaseResult `json:"result"`
}
