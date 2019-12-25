package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
)

type requestData struct {
	URL string `json:"url"`
}

type localFile struct {
	Path string `json:"path"`
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
	fmt.Printf("userID %v, StreamID %v\n", userID, streamID)
	reqData := new(requestData)
	err = c.Bind(reqData)
	if err != nil {
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
		err = s.DownloadFromURL(url, dstPath)
		if err != nil {
			s.logger.Error(err)
			s.ProcessErrCh <- err
			return err
		}

		return nil
	}(reqData.URL, dstPath)

	return c.NoContent(http.StatusCreated)
}

func (s *UploaderService) uploadFromFile(c echo.Context) error {
	streamID := c.Param("id")
	userID, err := s.authenticate(c)
	if err != nil {
		return err
	}
	fmt.Printf("userID %v, StreamID %v\n", userID, streamID)

	file, err := c.FormFile("file")

	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dstPath, err := s.getDestinationPath(file.Filename)
	if err != nil {
		s.logger.Error(err)
		s.ProcessErrCh <- err
		return err
	}
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.NoContent(http.StatusCreated)
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
