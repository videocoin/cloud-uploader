package downloader

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-uploader/datastore"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
)

var (
	ErrInvalidURL                 = errors.New("Invalid URL")
	ErrInvalidVideoFormat         = errors.New("Invalid video format")
	ErrInvalidVideoSize           = errors.New("Invalid video file size")
	ErrInvalidGDriveURL           = errors.New("Invalid Google Drive URL")
	ErrUnsupportedVideotypeFormat = errors.New("Unsupported stereo type")
	ErrFailedUpload               = errors.New("Failed to upload video from link. Check that the linked video exists and is publicly accessible")
)

type Downloader struct {
	dst       string
	gdriveKey string
	logger    *logrus.Entry
	ds        datastore.Datastore
	InputCh   chan *InputFile
	OutputCh  chan *OutputFile
}

func NewDownloader(ctx context.Context, dst string, opts ...Option) (*Downloader, error) {
	d := &Downloader{
		logger:    ctxlogrus.Extract(ctx).WithField("system", "downloader"),
		dst:       dst,
		gdriveKey: GDriveKeyFromContext(ctx),
		InputCh:   make(chan *InputFile, 1),
		OutputCh:  make(chan *OutputFile, 1),
	}

	for _, o := range opts {
		o(d)
	}

	return d, nil
}

func (d *Downloader) dispatch() {
	for f := range d.InputCh {
		go func(f *InputFile) {
			if f == nil {
				return
			}

			logger := d.logger.WithField("url", f.URL).WithField("stream_id", f.StreamID)
			logger.Info("downloading")

			ctx := ctxlogrus.ToContext(context.Background(), logger)

			outputFile, err := d.download(ctx, f)
			if err != nil {
				logger.WithError(err).Error("failed to download")
			}

			go func() {
				d.OutputCh <- outputFile
			}()
		}(f)
	}
}

func (d *Downloader) download(ctx context.Context, f *InputFile) (*OutputFile, error) {
	logger := ctxlogrus.Extract(ctx)

	u, err := url.Parse(f.URL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	q := u.Query()

	var output *OutputFile

	dstPath := f.GenDestPath(d.dst)
	dstFolder := path.Dir(dstPath)

	_, err = os.Stat(dstFolder)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		mkdirErr := os.MkdirAll(dstFolder, 0777)
		if mkdirErr != nil {
			return nil, mkdirErr
		}
	}

	f.DestPath = dstPath

	if strings.HasPrefix(u.Host, "drive.google") {
		logger.Info("downloading file from gdrive")

		gdriveID, err := parseGDriveFileID(f.URL)
		if err != nil {
			return nil, err
		}
		f.GDriveID = gdriveID

		output, err = d.downloadFileFromGdrive(f)
		if err != nil {
			return nil, err
		}
	} else {
		if strings.HasPrefix(u.Host, "www.dropbox") ||
			strings.HasPrefix(u.Host, "dropbox") {

			logger.Info("downloading file dropbox")

			q.Set("raw", "1")
			u.RawQuery = q.Encode()
			f.URL = u.String()
		} else {
			logger.Info("downloading general link")
		}

		output, err = d.downloadFile(f)
		if err != nil {
			return nil, err
		}
	}

	logger.WithField("dst", output.Path).Info("file has been downloaded")

	return output, nil
}

func (d *Downloader) downloadFile(f *InputFile) (*OutputFile, error) {
	resp, err := http.Get(f.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)

	dst, err := os.Create(f.DestPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if d.ds != nil {
		meta := &datastore.FileMeta{
			ID:   f.StreamID,
			Path: f.DestPath,
			Size: size,
		}
		err := d.ds.CreateFileMeta(context.Background(), meta)
		if err != nil {
			d.logger.
				WithField("stream_id", f.StreamID).
				WithError(err).
				Error("failed to create file meta")
		}
	}

	if _, err := io.Copy(dst, resp.Body); err != nil {
		return nil, err
	}

	out := &OutputFile{
		StreamID: f.StreamID,
		Path:     f.DestPath,
		Size:     size,
	}

	return out, nil
}

func (d *Downloader) downloadFileFromGdrive(f *InputFile) (*OutputFile, error) {
	ctx := context.Background()

	srv, err := drive.NewService(ctx, option.WithAPIKey(d.gdriveKey))
	if err != nil {
		return nil, err
	}

	_, err = srv.Files.Get(f.GDriveID).Fields("id,name,mimeType,size,webContentLink").Do()
	if err != nil {
		return nil, err
	}

	resp, err := srv.Files.Get(f.GDriveID).Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dst, err := os.Create(f.DestPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if d.ds != nil {
		meta := &datastore.FileMeta{
			ID:   f.StreamID,
			Path: f.DestPath,
			Size: resp.ContentLength,
		}
		err := d.ds.CreateFileMeta(ctx, meta)
		if err != nil {
			d.logger.
				WithField("stream_id", f.StreamID).
				WithError(err).
				Error("failed to create file meta")
		}
	}

	if _, err = io.Copy(dst, resp.Body); err != nil {
		return nil, err
	}

	out := &OutputFile{
		StreamID: f.StreamID,
		Path:     f.DestPath,
		Size:     resp.ContentLength,
	}

	return out, nil
}

func (d *Downloader) Start() error {
	_, err := os.Stat(d.dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		mkdirErr := os.MkdirAll(d.dst, 0777)
		if mkdirErr != nil {
			return mkdirErr
		}

		return err
	}

	d.dispatch()

	return nil
}

func (d *Downloader) Stop() error {
	close(d.InputCh)
	close(d.OutputCh)
	return nil
}

func (d *Downloader) Dst() string {
	return d.dst
}
