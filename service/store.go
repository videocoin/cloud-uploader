package service

import (
	"encoding/json"
)

type Record struct {
	Id   string
	Size int64
	Path string
}

func (s *UploaderService) CreateMetadataRecord(id string, size int64, path string) error {

	s.logger.Info("create metadata record")
	record := Record{
		Id:   id,
		Size: size,
		Path: path,
	}
	infoRaw, err := json.Marshal(record)
	if err != nil {
		return err
	}

	err = s.cli.Set(id, infoRaw, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (s *UploaderService) getMetadataRecord(id string) (*Record, error) {
	record := new(Record)
	recordRaw, err := s.cli.Get(id).Result()

	err = json.Unmarshal([]byte(recordRaw), record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s *UploaderService) clearMetadataRecord(id string) error {
	_, err := s.cli.Del(id).Result()
	if err != nil {
		return err
	}

	return nil
}
