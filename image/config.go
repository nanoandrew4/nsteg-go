package stegimg

import "image/png"

const (
	defaultChunkSizeMultiplier = 32 * 1024
)

type Config struct {
	LSBsToUse           byte
	ChunkSizeMultiplier int
	PngCompressionLevel png.CompressionLevel
}

func (c Config) populateUnsetConfigVars() {
	if c.LSBsToUse < 1 || c.LSBsToUse > 8 {
		c.LSBsToUse = 3
	}
	if c.ChunkSizeMultiplier < 1 {
		c.ChunkSizeMultiplier = defaultChunkSizeMultiplier
	}
}
