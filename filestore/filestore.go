package filestore

import (
	"errors"
)

var (
	ErrFileAlreadyExists = errors.New("File already exists")
	ErrFileNotFound      = errors.New("File not found")
	ErrFileTooLarge      = errors.New("File is too large")
	ErrFileCorrupted     = errors.New("File is corrupted, try storing it again")
	ErrChecksumFailed    = errors.New("Unable to store file: checksum verification failed")

	errKeysMissing = errors.New("Some keys are missing")
)

type Store interface {
	Store(filename string, contents []byte) error
	Retrieve(filename string) ([]byte, error)
	Delete(filename string) error
}
