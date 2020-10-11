package mock

import (
	"filestore"
	"fmt"

	log "github.com/sirupsen/logrus"
)

type mockStore struct {
	maxFileSize int
	files       map[string][]byte
}

func NewFilestore(maxFileSize int) filestore.Store {
	return mockStore{
		maxFileSize: maxFileSize,
		files:       map[string][]byte{},
	}
}

func (s mockStore) Store(filename string, contents []byte) error {
	size := len(contents)

	log.WithField("filename", filename).WithField("size", size).Debug("Storing file")

	if size > s.maxFileSize {
		return fmt.Errorf("%w: max file size is %d bytes", filestore.ErrFileTooLarge, s.maxFileSize)
	}

	if _, ok := s.files[filename]; ok {
		return filestore.ErrFileAlreadyExists
	}

	s.files[filename] = contents

	log.WithField("filename", filename).WithField("size", len(contents)).Info("Stored file")

	return nil
}

func (s mockStore) Retrieve(filename string) ([]byte, error) {
	log.WithField("filename", filename).Debug("Retrieving file")

	if contents, ok := s.files[filename]; ok {
		log.WithField("filename", filename).WithField("size", len(contents)).Info("Retrieved file")
		return contents, nil
	}

	return []byte{}, filestore.ErrFileNotFound
}

func (s mockStore) Delete(filename string) error {
	log.WithField("filename", filename).Debug("Deleting file")

	if _, ok := s.files[filename]; ok {
		delete(s.files, filename)
		log.WithField("filename", filename).Info("Deleted file")
		return nil
	}

	return filestore.ErrFileNotFound
}
