package downloader

import (
	"net/url"
	"strings"
)

func parseGDriveFileID(urlStr string) (string, error) {
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
