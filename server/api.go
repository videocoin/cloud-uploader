package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/videocoin/cloud-uploader/splitter"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	pstreamsv1 "github.com/videocoin/cloud-api/streams/private/v1"
	streamsv1 "github.com/videocoin/cloud-api/streams/v1"
	"github.com/videocoin/cloud-uploader/downloader"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MIMEVideoMP4       = "video/mp4"
	MIMEVideoQuickTime = "video/quicktime"
)

func (s *Server) getHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]bool{"alive": true})
}

func (s *Server) uploadFromURL(c echo.Context) error {
	streamID := c.Param("id")

	logger := s.logger.WithField("stream_id", streamID)

	userID, err := s.authenticate(c)
	if err != nil {
		logger.WithError(err).Warningf("failed to auth")
		return err
	}

	err, httpErr := s.validate(c.Request().Context(), streamID, userID)
	if httpErr != nil {
		if err != nil {
			logger.WithError(err).Error("failed to validate")
		}
		return httpErr
	}

	reqData := new(requestData)
	err = c.Bind(reqData)
	if err != nil {
		logger.WithError(err).Error("failed to bind data")
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid data")
	}

	logger = logger.WithField("url", reqData.URL)
	logger.Info("uplodaing from url")

	s.downloader.InputCh <- &downloader.InputFile{
		StreamID: streamID,
		URL:      reqData.URL,
	}

	return c.NoContent(http.StatusCreated)
}

func (s *Server) checkUploadFromURL(c echo.Context) error {
	streamID := c.Param("id")

	logger := s.logger.WithField("stream_id", streamID)

	userID, err := s.authenticate(c)
	if err != nil {
		logger.WithError(err).Warningf("failed to auth")
		return err
	}

	httpErr, err := s.validate(c.Request().Context(), streamID, userID)
	if httpErr != nil {
		if err != nil {
			logger.WithError(err).Error("failed to validate")
		}
		return httpErr
	}

	resp := &progressResponse{Progress: 0}

	if s.ds != nil {
		meta, err := s.ds.GetFileMeta(c.Request().Context(), streamID)
		if err != nil {
			logger.WithError(err).Error("failed to get file meta")
		}
		if meta != nil {
			f, err := os.Stat(meta.Path)
			if err == nil {
				resp.Progress = f.Size() * 100 / meta.Size
			}
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) uploadFromFile(c echo.Context) error {
	streamID := c.Param("id")

	logger := s.logger.WithField("stream_id", streamID)

	userID, err := s.authenticate(c)
	if err != nil {
		logger.WithError(err).Warningf("failed to auth")
		return err
	}

	httpErr, err := s.validate(c.Request().Context(), streamID, userID)
	if httpErr != nil {
		if err != nil {
			logger.WithError(err).Error("failed to validate")
		}
		return httpErr
	}

	f, err := c.FormFile("file")
	if err != nil {
		logger.Errorf("failed to form file: %s", err)
		return err
	}

	src, err := f.Open()
	if err != nil {
		logger.Errorf("failed to file open: %s", err)
		return err
	}
	defer src.Close()

	dstPath := fmt.Sprintf("%s/%s/%s", s.downloader.Dst(), streamID, f.Filename)
	dstFolder := path.Dir(dstPath)

	_, err = os.Stat(dstFolder)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.WithError(err).Error("failed to stat destination folder")
			return echo.ErrInternalServerError
		}

		mkdirErr := os.MkdirAll(dstFolder, 0777)
		if mkdirErr != nil {
			logger.WithError(mkdirErr).Error("failed to create destination folder")
			return echo.ErrInternalServerError
		}
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		logger.WithError(err).Error("failed to create destination file")
		return echo.ErrInternalServerError
	}
	defer dst.Close()

	logger.Info("uploading")

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	if s.splitter != nil {
		mediaFile := &splitter.MediaFile{
			StreamID: streamID,
			Path:     dstPath,
		}

		go func(mf *splitter.MediaFile) {
			s.splitter.InputCh <- mf
		}(mediaFile)
	}

	return c.NoContent(http.StatusCreated)
}

func (s *Server) validate(ctx context.Context, streamID, userID string) (*echo.HTTPError, error) {
	req := &pstreamsv1.StreamRequest{Id: streamID}
	stream, err := s.sc.Streams.Get(ctx, req)
	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.NotFound {
				return echo.ErrNotFound, err
			}
		}

		return echo.ErrInternalServerError, err
	}

	if stream.UserID != userID {
		return nil, echo.ErrNotFound
	}

	if stream.Status != streamsv1.StreamStatusPrepared {
		return echo.NewHTTPError(http.StatusBadRequest, "Stream isn't prepeared"),
			fmt.Errorf("wrong stream status: %s", stream.Status.String())
	}

	return nil, nil
}

func (s *Server) authenticate(ctx echo.Context) (string, error) {
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
