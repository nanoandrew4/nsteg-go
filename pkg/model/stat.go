package model

import (
	"time"
)

type EncodeStats struct {
	Setup               time.Duration `json:"setup"`
	DataEncoding        time.Duration `json:"data_encoding"`
	OutputImageEncoding time.Duration `json:"output_image_encoding"`
}

type DecodeStats struct {
	DataDecoding time.Duration `json:"data_decoding"`
}
