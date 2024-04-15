package model

import "io"

type InputFile struct {
	Name    string
	Content io.Reader
	Size    int64
}

type OutputFile struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}
