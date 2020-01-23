package service

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
	splitterv1 "github.com/videocoin/cloud-api/splitter/v1"
	pstreamsv1 "github.com/videocoin/cloud-api/streams/private/v1"
	streamsv1 "github.com/videocoin/cloud-api/streams/v1"
)

const (
	MIMEVideoMP4       = "video/mp4"
	MIMEVideoQuickTime = "video/quicktime"
)

type requestData struct {
	URL string `json:"url"`
}

type localFile struct {
	Path string `json:"path"`
}

type ProgressResponse struct {
	Progress int32 `json:"progress"`
}

func (s *UploaderService) getHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]bool{"alive": true})
}

func (s *UploaderService) uploadFromURL(c echo.Context) error {
	streamID := c.Param("id")
	userID, err := s.authenticate(c)
	if err != nil {
		return err
	}

	err = s.validate(c, userID)
	if err != nil {
		s.logger.Warningf("failed to validate: %s", err)
		return err
	}

	reqData := new(requestData)
	err = c.Bind(reqData)
	if err != nil {
		s.logger.Errorf("failed to bind data: %s", err)
		return err
	}

	filename := getFilenameFromURL(reqData.URL)
	dstPath, err := s.getDestinationPath(filename)
	if err != nil {
		s.logger.Error(err)
		s.ProcessErrCh <- err
		return err
	}

	go func(url string, dstPath string) error {
		err = s.DownloadFromURL(streamID, url, dstPath)
		if err != nil {
			s.logger.Error(err)
			s.ProcessErrCh <- err
			return err
		}
		s.notifySplitter(streamID, dstPath)
		return nil
	}(reqData.URL, dstPath)

	return c.NoContent(http.StatusCreated)
}

func (s *UploaderService) checkUploadFromURL(c echo.Context) error {
	streamID := c.Param("id")
	userID, err := s.authenticate(c)
	if err != nil {
		return err
	}

	err = s.validate(c, userID)
	if err != nil {
		s.logger.Warningf("failed to validate: %s", err)
		return err
	}
	record, err := s.getMetadataRecord(streamID)
	if err != nil {
		return c.JSON(http.StatusNotFound, record)
	}
	fi, err := os.Stat(record.Path)
	if err != nil {
		return err
	}
	offset := fi.Size()

	progressResponse := ProgressResponse{
		Progress: int32(offset * 100 / record.Size),
	}
	return c.JSON(http.StatusOK, progressResponse)
}

func (s *UploaderService) uploadFromFile(c echo.Context) error {
	streamID := c.Param("id")

	s.logger.Info("authentication")

	userID, err := s.authenticate(c)
	if err != nil {
		s.logger.Warningf("failed to auth: %s", err)
		return err
	}

	s.logger.Info("validating")

	err = s.validate(c, userID)
	if err != nil {
		s.logger.Warningf("failed to validate: %s", err)
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		s.logger.Errorf("failed to form file: %s", err)
		return err
	}

	src, err := file.Open()
	if err != nil {
		s.logger.Errorf("failed to file open: %s", err)
		return err
	}
	defer src.Close()

	dstPath, err := s.getDestinationPath(file.Filename)
	if err != nil {
		s.logger.Errorf("failed to get destination path: %s", err)
		return err
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		s.logger.Errorf("failed to create file: %s", err)
		return err
	}
	defer dst.Close()

	logger := s.logger.WithField("id", streamID)

	req := &pstreamsv1.StreamRequest{Id: streamID}
	streamResp, err := s.streams.Get(context.Background(), req)
	if err != nil {
		logger.Errorf("failed to get stream: %s", err)
		return err
	}

	if streamResp.Status != streamsv1.StreamStatusPrepared {
		logger.Errorf("wrong stream status: %s", streamResp.Status.String())
		return echo.NewHTTPError(http.StatusBadRequest, "Stream isn't prepeared")
	}

	logger.Info("uploading")

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	logger.Info("send to split")

	s.notifySplitter(streamResp.ID, dstPath)

	return c.NoContent(http.StatusCreated)
}

func (s *UploaderService) validate(ctx echo.Context, userID string) error {
	streamID := ctx.Param("id")

	stream, err := s.streams.Get(context.Background(), &pstreamsv1.StreamRequest{Id: streamID})
	if err != nil {
		return err
	}

	if stream.UserID != userID {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	return nil
}

func (s *UploaderService) authenticate(ctx echo.Context) (string, error) {
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	span, _ := opentracing.StartSpanFromContext(ctx.Request().Context(), "authenticate")
	defer span.Finish()

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Please provide valid credentials")
	}

	return userID, nil
}

func (s *UploaderService) notifySplitter(streamID string, filepath string) error {
	_, err := s.splitter.Split(context.Background(), &splitterv1.SplitRequest{StreamID: streamID, Filepath: filepath})
	if err != nil {
		return err
	}

	return nil
}

func (s *UploaderService) getDestinationPath(filename string) (string, error) {
	uploadID := uuid.New().String()
	_, err := os.Stat(filepath.Join(s.config.DownloadDir, uploadID))
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}

		//mkdirErr := os.MkdirAll(filepath.Join(s.config.DownloadDir, uploadID), os.ModeDir)
		mkdirErr := os.MkdirAll(filepath.Join(s.config.DownloadDir, uploadID), 0777)
		if mkdirErr != nil {
			return "", err
		}
	}

	dstPath := filepath.Join(s.config.DownloadDir, uploadID, filename)
	return dstPath, nil
}
