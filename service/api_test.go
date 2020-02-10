// +build integration

package service

import (
	"io"
	"os"
	"fmt"
	"bytes"
	"strings"
	"testing"
	"mime/multipart"
	"net/http"
	"net/http/httptest"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/videocoin/cloud-pkg/logger"
	"github.com/videocoin/cloud-uploader/mock"
)

var (
	ServiceName string = "uploader"
	Version     string = "test"
)

type APITestSuite struct {
	suite.Suite
	svc *UploaderService
}

func (suite *APITestSuite) SetupSuite() {
	logger.Init(ServiceName, Version)

	log := logrus.NewEntry(logrus.New())
	config := &Config{
		Name:    "uploader",
		Version: "test",
	}

	err := envconfig.Process(ServiceName, config)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	config.Logger = log
	privateStreamManager := new(mock.MockPrivateStreamManager)
	splitterManager := new(mock.MockSplitterManager)
	suite.svc, err = NewService(config, privateStreamManager, splitterManager)
	require.NoError(suite.T(), err)

	suite.svc.route()
	require.NoError(suite.T(), err)
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

func (suite *APITestSuite) TestUpload_FromFile() {
	filepath := "testdata/small.mp4"
	file, err := os.Open(filepath)
	defer file.Close()
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	part, err := writer.CreateFormFile("file", "small.mp4")
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	size, err := io.Copy(part, file)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}
	fmt.Fprintf(os.Stdout, "Copied %v bytes for uploading...\n", size)
	writer.Close()
	req := httptest.NewRequest(echo.POST, fmt.Sprintf("/api/v1/upload/local/%s", mock.STREAM_ID), &buffer)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mock.GetAuthToken(suite.svc.config.AuthTokenSecret)))
	rec := httptest.NewRecorder()
	suite.svc.api.ServeHTTP(rec, req)
	resp := rec.Result()

	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	assert.Empty(suite.T(), rec.Body.String())

}

func (suite *APITestSuite) TestUpload_FromURL_WithGeneralURL() {
	requestJSON := `{"url":"http://techslides.com/demos/sample-videos/small.mp4"}`
	testUploadFromURL(requestJSON, suite)
}

func (suite *APITestSuite) TestUpload_FromURL_WithDropboxURL() {
	requestJSON := `{"url":"https://www.dropbox.com/s/ct6ibw2nucwkpii/big_buck_bunny_720p_1mb.mp4?dl=0"}`
	testUploadFromURL(requestJSON, suite)
}

func (suite *APITestSuite) TestUpload_FromURL_WithGDriveURL() {
	requestJSON := `{"url":"https://drive.google.com/open?id=0B_gJUUPAyx4ARUYxUFVHa05weHc"}`
	testUploadFromURL(requestJSON, suite)
}

func testUploadFromURL(requestJSON string, suite *APITestSuite) {
	body := strings.NewReader(requestJSON)
	req := httptest.NewRequest(echo.POST, fmt.Sprintf("/api/v1/upload/url/%s", mock.STREAM_ID), body)

	testSetupRequestHeader(req, mock.GetAuthToken(suite.svc.config.AuthTokenSecret))

	rec := httptest.NewRecorder()
	suite.svc.api.ServeHTTP(rec, req)
	resp := rec.Result()

	require.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	assert.Empty(suite.T(), rec.Body.String())
}

func testSetupRequestHeader(req *http.Request, authToken string) {
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
}
