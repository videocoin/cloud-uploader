package downloader

import (
	"fmt"
	"path"
	"strings"
)

type InputFile struct {
	StreamID string
	URL      string
	GDriveID string
	DestPath string
}

func (f *InputFile) Filename() string {
	parts := strings.Split(f.URL, "?")
	return fmt.Sprintf("%s%s", f.StreamID, path.Ext(parts[0]))
}

func (f *InputFile) GenDestPath(base string) string {
	return fmt.Sprintf("%s/%s/%s", base, f.StreamID, f.Filename())
}

type OutputFile struct {
	StreamID string
	Path     string
	Size     int64
	Error    error
}
