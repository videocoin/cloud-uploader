package test

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"io"
// 	"mime/multipart"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"path"
// 	"strings"
// 	"testing"

// 	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
// 	"github.com/kelseyhightower/envconfig"
// 	"github.com/labstack/echo/v4"
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"
// 	clientv1 "github.com/videocoin/cloud-api/client/v1"
// 	"github.com/videocoin/cloud-uploader/datastore"
// 	"github.com/videocoin/cloud-uploader/downloader"
// 	"github.com/videocoin/cloud-uploader/mock"
// 	"github.com/videocoin/cloud-uploader/server"
// 	"github.com/videocoin/cloud-uploader/service"
// 	"github.com/videocoin/cloud-uploader/splitter"
// )

// var (
// 	ServiceName string = "uploader"
// 	Version     string = "test"
// )

// type APITestSuite struct {
// 	suite.Suite
// 	logger     *logrus.Entry
// 	cfg        *service.Config
// 	srv        *server.Server
// 	ds         datastore.Datastore
// 	downloader *downloader.Downloader
// 	splitter   *splitter.Splitter
// }

// func (suite *APITestSuite) SetupSuite() {
// 	logger := logrus.NewEntry(logrus.New())
// 	suite.logger = logger

// 	ctx := ctxlogrus.ToContext(context.Background(), logger)

// 	cfg := &service.Config{
// 		Name:    ServiceName,
// 		Version: Version,
// 	}
// 	err := envconfig.Process(ServiceName, cfg)
// 	require.NoError(suite.T(), err)

// 	cfg.AuthTokenSecret = "secret"
// 	cfg.DownloadDir = "/tmp/hls"

// 	suite.cfg = cfg

// 	suite.ds = mock.NewDatastore()
// 	downloader, err := downloader.NewDownloader(ctx, cfg.DownloadDir, downloader.WithDatastore(suite.ds))
// 	require.NoError(suite.T(), err)
// 	suite.downloader = downloader
// 	go func() {
// 		err := suite.downloader.Start()
// 		require.NoError(suite.T(), err)
// 	}()

// 	splitter, err := splitter.NewSplitter(ctx, splitter.WithOutputDir(cfg.DownloadDir))
// 	require.NoError(suite.T(), err)
// 	suite.splitter = splitter
// 	go func() {
// 		err := suite.splitter.Start()
// 		require.NoError(suite.T(), err)
// 	}()

// 	sc := &clientv1.ServiceClient{
// 		Streams: &mock.PrivateStreamManager{ID: mock.StreamID},
// 	}

// 	suite.srv, err = server.NewServer(
// 		ctx,
// 		server.WithAddr(cfg.Addr),
// 		server.WithAuthTokenSecret(cfg.AuthTokenSecret),
// 		server.WithDownloader(downloader),
// 		server.WithServiceClient(sc),
// 		server.WithDatastore(suite.ds),
// 	)
// 	require.NoError(suite.T(), err)
// }

// func (suite *APITestSuite) TearDownSuite() {
// 	err := suite.downloader.Stop()
// 	require.NoError(suite.T(), err)

// 	err = suite.splitter.Stop()
// 	require.NoError(suite.T(), err)
// }

// func TestAPITestSuite(t *testing.T) {
// 	suite.Run(t, new(APITestSuite))
// }

// func (suite *APITestSuite) TestUpload_FromFile() {
// 	filepath := "testdata/small.mp4"
// 	f, err := os.Open(filepath)
// 	require.NoError(suite.T(), err)
// 	defer f.Close()

// 	var buffer bytes.Buffer
// 	writer := multipart.NewWriter(&buffer)
// 	part, err := writer.CreateFormFile("file", "small.mp4")
// 	require.NoError(suite.T(), err)

// 	_, err = io.Copy(part, f)
// 	require.NoError(suite.T(), err)

// 	writer.Close()

// 	req := httptest.NewRequest(echo.POST, fmt.Sprintf("/api/v1/upload/local/%s", mock.StreamID), &buffer)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mock.GetAuthToken(suite.cfg.AuthTokenSecret)))

// 	rec := httptest.NewRecorder()
// 	suite.srv.E().ServeHTTP(rec, req)
// 	resp := rec.Result()

// 	dstPath := fmt.Sprintf(
// 		"%s/%s/%s",
// 		suite.downloader.Dst(),
// 		mock.StreamID,
// 		path.Base(filepath),
// 	)

// 	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
// 	assert.Empty(suite.T(), rec.Body.String())

// 	_, err = os.Stat(dstPath)
// 	require.NoError(suite.T(), err)

// 	suite.splitter.InputCh <- &splitter.MediaFile{
// 		StreamID: mock.StreamID,
// 		Path:     dstPath,
// 	}

// 	mediaFile := <-suite.splitter.OutputCh
// 	require.NoError(suite.T(), mediaFile.Error)
// }

// func (suite *APITestSuite) TestUpload_FromURL_WithGeneralURL() {
// 	requestJSON := `{"url":"http://techslides.com/demos/sample-videos/small.mp4"}`
// 	testUploadFromURL(requestJSON, suite)
// }

// func (suite *APITestSuite) TestUpload_FromURL_WithDropboxURL() {
// 	requestJSON := `{"url":"https://www.dropbox.com/s/ct6ibw2nucwkpii/big_buck_bunny_720p_1mb.mp4?dl=0"}`
// 	testUploadFromURL(requestJSON, suite)
// }

// func (suite *APITestSuite) TestUpload_FromURL_WithGDriveURL() {
// 	requestJSON := `{"url":"https://drive.google.com/open?id=0B_gJUUPAyx4ARUYxUFVHa05weHc"}`
// 	testUploadFromURL(requestJSON, suite)
// }

// func testUploadFromURL(requestJSON string, suite *APITestSuite) {
// 	body := strings.NewReader(requestJSON)
// 	req := httptest.NewRequest(echo.POST, fmt.Sprintf("/api/v1/upload/url/%s", mock.StreamID), body)

// 	testSetupRequestHeader(req, mock.GetAuthToken(suite.cfg.AuthTokenSecret))

// 	rec := httptest.NewRecorder()
// 	suite.srv.E().ServeHTTP(rec, req)
// 	resp := rec.Result()

// 	outputFile := <-suite.downloader.OutputCh
// 	require.NotNil(suite.T(), outputFile)

// 	require.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
// 	assert.Empty(suite.T(), rec.Body.String())

// 	_, err := os.Stat(outputFile.Path)
// 	require.NoError(suite.T(), err)

// 	emptyCtx := context.Background()
// 	meta, err := suite.ds.GetFileMeta(emptyCtx, mock.StreamID)
// 	require.NoError(suite.T(), err)
// 	require.NotNil(suite.T(), meta)
// 	assert.Equal(suite.T(), meta.ID, mock.StreamID)

// 	suite.splitter.InputCh <- &splitter.MediaFile{
// 		StreamID: outputFile.StreamID,
// 		Path:     outputFile.Path,
// 	}

// 	mediaFile := <-suite.splitter.OutputCh
// 	require.NoError(suite.T(), mediaFile.Error)
// }

// func testSetupRequestHeader(req *http.Request, bearer string) {
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearer))
// }
