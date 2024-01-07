package internal

import "io"

type FileToHide struct {
	Name    string
	Content io.Reader
	Size    int64
}
