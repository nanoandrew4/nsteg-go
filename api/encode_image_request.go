package api

type EncodeImageRequest struct {
	LsbsToUse     byte         `json:"lsbs_to_use"`
	ImageToEncode []byte       `json:"image_to_encode"`
	FilesToHide   []FileToHide `json:"files_to_hide"`
}
