package service

import (
	"net/url"
	"path/filepath"
	"strings"
)

func getFilenameFromURL(urlStr string) string {
	_, filename := filepath.Split(urlStr)
	s1 := strings.Split(filename, "?")
	s2 := strings.Split(s1[0], ".")
	s3, _ := url.PathUnescape(strings.Join(s2, "."))
	return s3
}

func getGDriveFileID(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", ErrInvalidURL
	}

	q := u.Query()
	if q.Get("id") != "" {
		return q.Get("id"), nil
	}

	paths := strings.Split(u.Path, "/")
	if len(paths) >= 4 {
		return paths[3], nil
	}

	return "", ErrInvalidGDriveURL
}
