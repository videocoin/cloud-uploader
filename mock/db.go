package mock

import (
	"context"

	"github.com/videocoin/cloud-uploader/datastore"
)

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
