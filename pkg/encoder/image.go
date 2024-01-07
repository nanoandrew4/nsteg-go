package encoder

import (
	"image"
	"nsteg/internal"
	"nsteg/internal/encoder"
)

func NewImageEncoder(img *image.RGBA, config internal.ImageEncodeConfig) *encoder.ImageEncoder {
	return encoder.NewImageEncoder(img, config)
}
