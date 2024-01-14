package config

import "image/png"

const (
	DefaultChunkSizeMultiplier = 32 * 1024
)

type ImageEncodeConfig struct {
	LSBsToUse           byte
	ChunkSizeMultiplier int
	PngCompressionLevel png.CompressionLevel
}

func (c ImageEncodeConfig) PopulateUnsetConfigVars() {
	if c.LSBsToUse < 1 || c.LSBsToUse > 8 {
		c.LSBsToUse = 3
	}
	if c.ChunkSizeMultiplier < 1 {
		c.ChunkSizeMultiplier = DefaultChunkSizeMultiplier
	}
}
