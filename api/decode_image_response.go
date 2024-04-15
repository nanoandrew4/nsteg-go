package api

import "nsteg/pkg/model"

type DecodeImageResponse struct {
	DecodedFiles []model.OutputFile `json:"decoded_files"`
}
