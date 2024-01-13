package api

type FileToHide struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}
