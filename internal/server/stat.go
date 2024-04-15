package server

import (
	"nsteg/pkg/model"
)

type humanizedEncodeStats struct {
	model.EncodeStats
	SetupHuman               string `json:"setup_human"`
	DataEncodingHuman        string `json:"data_encoding_human"`
	OutputImageEncodingHuman string `json:"output_image_encoding_human"`
}

type humanizedDecodeStats struct {
	model.DecodeStats
	DataDecodingHuman string `json:"data_decoding_human"`
}

func toHumanizedEncodeStats(encodeStats model.EncodeStats) humanizedEncodeStats {
	return humanizedEncodeStats{
		EncodeStats:              encodeStats,
		SetupHuman:               encodeStats.Setup.String(),
		DataEncodingHuman:        encodeStats.DataEncoding.String(),
		OutputImageEncodingHuman: encodeStats.OutputImageEncoding.String(),
	}
}

func toHumanizedDecodeStats(decodeStats model.DecodeStats) humanizedDecodeStats {
	return humanizedDecodeStats{
		DecodeStats:       decodeStats,
		DataDecodingHuman: decodeStats.DataDecoding.String(),
	}
}
