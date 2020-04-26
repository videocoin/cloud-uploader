package mock

import (
	"context"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	types "github.com/gogo/protobuf/types"
	splitterv1 "github.com/videocoin/cloud-api/splitter/v1"
	pstreamsv1 "github.com/videocoin/cloud-api/streams/private/v1"
	streamsv1 "github.com/videocoin/cloud-api/streams/v1"
	usersv1 "github.com/videocoin/cloud-api/users/v1"
	"github.com/videocoin/cloud-pkg/auth"
	"github.com/videocoin/cloud-uploader/datastore"
	"google.golang.org/grpc"
)

const UserID = "12b1876f-341f-41b0-833f-5312f1e9c308"
const StreamID = "cdc1816b-0be8-44a6-80c3-3e43fbd441ee"

func GetAuthToken(authTokenSecret string) string {
	claims := auth.ExtendedClaims{
		Type: auth.TokenType(usersv1.TokenTypeRegular),
		StandardClaims: jwt.StandardClaims{
			Subject:   UserID,
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

type PrivateStreamManager struct {
	ID string
}

func (sm *PrivateStreamManager) Get(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID:     sm.ID,
		UserID: UserID,
		Status: streamsv1.StreamStatusPrepared,
	}
	return &stream, nil
}

func (sm *PrivateStreamManager) Publish(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.ID,
	}
	return &stream, nil
}

func (sm *PrivateStreamManager) PublishDone(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.ID,
	}
	return &stream, nil
}

func (sm *PrivateStreamManager) Run(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.ID,
	}
	return &stream, nil
}

func (sm *PrivateStreamManager) Stop(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.ID,
	}
	return &stream, nil
}

func (sm *PrivateStreamManager) Complete(
	context.Context, *pstreamsv1.StreamRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.ID,
	}
	return &stream, nil
}

func (sm *PrivateStreamManager) UpdateStatus(
	context.Context, *pstreamsv1.UpdateStatusRequest, ...grpc.CallOption,
) (*pstreamsv1.StreamResponse, error) {
	stream := pstreamsv1.StreamResponse{
		ID: sm.ID,
	}
	return &stream, nil
}

type SplitterManager struct{}

func (sm *SplitterManager) Split(
	context.Context, *splitterv1.SplitRequest, ...grpc.CallOption,
) (*types.Empty, error) {
	return nil, nil
}

type Datastore struct {
	data map[string]interface{}
}

func NewDatastore() *Datastore {
	return &Datastore{data: map[string]interface{}{}}
}

func (ds *Datastore) Start() error {
	return nil
}

func (ds *Datastore) Stop() error {
	return nil
}

func (ds *Datastore) CreateFileMeta(ctx context.Context, meta *datastore.FileMeta) error {
	ds.data[meta.ID] = meta
	return nil
}

func (ds *Datastore) GetFileMeta(ctx context.Context, id string) (*datastore.FileMeta, error) {
	if meta, ok := ds.data[id]; ok {
		return meta.(*datastore.FileMeta), nil
	}
	return nil, nil
}
