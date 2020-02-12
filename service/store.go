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
	record_raw, err := s.cli.Get(id).Result()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(record_raw), record)
	if err != nil {
		return nil, err
	}

	return record, nil
}