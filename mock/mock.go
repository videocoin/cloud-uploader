package mock

import (
	"context"
	jwt "github.com/dgrijalva/jwt-go"
	types "github.com/gogo/protobuf/types"
	splitterv1 "github.com/videocoin/cloud-api/splitter/v1"
	pstreamsv1 "github.com/videocoin/cloud-api/streams/private/v1"
	streamsv1 "github.com/videocoin/cloud-api/streams/v1"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-pkg/auth"
	"google.golang.org/grpc"
	"time"
)

const USER_ID = "12b1876f-341f-41b0-833f-5312f1e9c308"
const STREAM_ID = "cdc1816b-0be8-44a6-80c3-3e43fbd441ee"

func GetAuthToken(authTokenSecret string) string {
	claims := auth.ExtendedClaims{
		Type: auth.TokenType(usersv1.TokenTypeRegular),
		StandardClaims: jwt.StandardClaims{
			Subject:   USER_ID,
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	st, err := token.SignedString([]byte(authTokenSecret))
	if err != nil {
		return ""
	}
	return st
}

type MockPrivateStreamManager struct {
	id string
}

func (sm *MockPrivateStreamManager) Get(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID:     sm.id,
		UserID: USER_ID,
		Status: streamsv1.StreamStatusPrepared,
	}
	return &stream, nil
}

func (sm *MockPrivateStreamManager) Publish(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.id,
	}
	return &stream, nil
}

func (sm *MockPrivateStreamManager) PublishDone(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.id,
	}
	return &stream, nil
}

func (sm *MockPrivateStreamManager) Run(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.id,
	}
	return &stream, nil
}

func (sm *MockPrivateStreamManager) Stop(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.id,
	}
	return &stream, nil
}

func (sm *MockPrivateStreamManager) UpdateStatus(
	context.Context, *pstreamsv1.UpdateStatusRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.id,
	}
	return &stream, nil
}

type MockSplitterManager struct {}

func (sm *MockSplitterManager) Split(
	context.Context, *splitterv1.SplitRequest, ...grpc.CallOption,
) (*types.Empty, error) {
	return nil, nil
}
