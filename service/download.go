package service

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi/transport"
)

var (
	ErrInvalidURL                 = errors.New("Invalid URL")
	ErrInvalidVideoFormat         = errors.New("Invalid video format")
	ErrInvalidVideoSize           = errors.New("Invalid video file size")
	ErrInvalidGDriveURL           = errors.New("Invalid Google Drive URL")
	ErrUnsupportedVideotypeFormat = errors.New("Unsupported stereo type")
	ErrFailedUpload               = errors.New("Failed to upload video from link. Check that the linked video exists and is publicly accessible.")
)

func (s *UploaderService) DownloadFromURL(streamID, urlStr, dstPath string) error {

	s.logger.Info("downloading file")

	u, err := url.Parse(urlStr)
	if err != nil {
		return ErrInvalidURL
	}

	q := u.Query()

	if strings.HasPrefix(u.Host, "drive.google") {
		GDriveID, err := getGDriveFileID(urlStr)
		err = s.downloadGdriveFile(streamID, GDriveID, dstPath)
		if err != nil {
			return err
		}
	} else {
		if strings.HasPrefix(u.Host, "www.dropbox") ||
			strings.HasPrefix(u.Host, "dropbox") {

			q.Set("raw", "1")
			u.RawQuery = q.Encode()
			urlStr = u.String()
		}

		err := s.downloadBaseFile(streamID, urlStr, dstPath)
		if err != nil {
			return err
		}
	}

	s.logger.Info("file has been downloaded")

	return nil
}

func (s *UploaderService) downloadBaseFile(streamID, urlStr, dstPath string) error {
	resp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	s.CreateMetadataRecord(streamID, int64(size), dstPath)

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	done := make(chan bool)

	if _, err := io.Copy(dst, resp.Body); err != nil {
		return err
	}
	done <- true

	return nil
}

func (s *UploaderService) downloadGdriveFile(streamID, gdriveId, dstPath string) error {
	srv, err := drive.New(&http.Client{
		Transport: &transport.APIKey{Key: s.config.GDriveKey},
	})
	if err != nil {
		return err
	}

	r, err := srv.Files.Get(gdriveId).Fields("id,name,mimeType,size,webContentLink").Do()
	if err != nil {
		return err
	}
	s.CreateMetadataRecord(streamID, int64(r.Size), dstPath)

	resp, err := srv.Files.Get(gdriveId).Download()
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	done := make(chan bool)
	if _, err = io.Copy(dst, resp.Body); err != nil {
		return err
	}
	done <- true

	return nil
}
