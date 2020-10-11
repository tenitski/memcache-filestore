package filestore

import (
	"crypto/md5"
	"encoding/hex"
	"filestore/client"
	"fmt"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	log "github.com/sirupsen/logrus"
)

const defaultChunkSize = 1024 * 1024        // 1MB
const defaultMaxFileSize = 50 * 1024 * 1024 // 50MB
const keyPrefix = "filestore:"

type memcacheStore struct {
	client      client.Memcache
	chunkSize   int
	maxFileSize int
}

type MemcacheConfig struct {
	Timeout     time.Duration
	ChunkSize   int
	MaxFileSize int
}

func NewMemcache(server string, config MemcacheConfig) Store {
	c := memcache.New(server)
	if config.Timeout != 0 {
		c.Timeout = config.Timeout
	}

	return NewMemcacheWithClient(c, config)
}

func NewMemcacheWithClient(client client.Memcache, config MemcacheConfig) Store {
	chunkSize := config.ChunkSize
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}

	maxFileSize := config.MaxFileSize
	if maxFileSize <= 0 {
		maxFileSize = defaultMaxFileSize
	}

	return &memcacheStore{
		client:      client,
		chunkSize:   chunkSize,
		maxFileSize: maxFileSize,
	}
}

func (s memcacheStore) Store(filename string, contents []byte) error {
	size := len(contents)

	log.WithField("filename", filename).WithField("size", size).Debug("Storing file")

	if size > s.maxFileSize {
		return fmt.Errorf("%w: max file size is %d bytes", ErrFileTooLarge, s.maxFileSize)
	}

	metadataKey := buildKey(filename)

	// Check if the file already exists
	_, err := s.getKey(metadataKey)
	if err == nil {
		return ErrFileAlreadyExists
	} else if err != memcache.ErrCacheMiss {
		return fmt.Errorf("Unable to store file: %w", err)
	}

	totalChunks := size / s.chunkSize
	if size%s.chunkSize > 0 {
		totalChunks++
	}

	// Create metadata key
	err = s.setKey(metadataKey, []byte(strconv.Itoa(totalChunks)))
	if err != nil {
		return fmt.Errorf("Unable to store file: %w", err)
	}

	// Create keys for each chunk
	index := 0
	for i := 0; i < size; i += s.chunkSize {
		end := i + s.chunkSize
		if end > size {
			end = size
		}

		chunkKey := buildChunkKey(filename, index)
		chunk := contents[i:end]

		err := s.setKey(chunkKey, chunk)
		if err != nil {
			purgeErr := s.purgeFile(filename, totalChunks)
			if purgeErr != nil {
				log.WithField("filename", filename).
					WithError(purgeErr).
					Error("Unable to cleanup file after storing failed")
			}

			return fmt.Errorf("Unable to store file: %w", err)
		}

		index++
	}

	// Check if file was stored completely
	// This is a naive and expensive way to do it
	// Memcache may offer a better way to verify if a key exists without actually retrieving it
	storedContents, err := s.Retrieve(filename)
	if err != nil {
		purgeErr := s.purgeFile(filename, totalChunks)
		if purgeErr != nil {
			log.WithField("filename", filename).
				WithError(purgeErr).
				Error("Unable to cleanup file after storing failed")
		}

		return fmt.Errorf("Unable to store file: %w", err)
	}

	if checksum(contents) != checksum(storedContents) {
		return ErrChecksumFailed
	}

	log.WithField("filename", filename).WithField("size", len(storedContents)).Info("Stored file")

	return nil
}

func (s memcacheStore) Retrieve(filename string) ([]byte, error) {
	log.WithField("filename", filename).Debug("Retrieving file")

	metadataKey := buildKey(filename)

	metadata, err := s.getKey(metadataKey)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return []byte{}, ErrFileNotFound
		}

		return []byte{}, fmt.Errorf("Unable to retrieve file: %w", err)
	}

	totalChunks, _ := strconv.Atoi(string(metadata)) // skipping error checking, will trust metadata not to have anything funny

	allKeys := []string{}
	for i := 0; i < totalChunks; i++ {
		chunkKey := buildChunkKey(filename, i)
		allKeys = append(allKeys, chunkKey)
	}

	values, err := s.getKeys(allKeys)
	if err != nil {
		if err == errKeysMissing {
			// There are less chunks than we expected, file is corrupted
			return []byte{}, ErrFileCorrupted
		}

		return []byte{}, fmt.Errorf("Unable to retrieve file: %w", err)
	}

	// Iterate over all returned chunks and build up file contents
	contents := []byte{}
	for i := 0; i < totalChunks; i++ {
		contents = append(contents, values[allKeys[i]]...)
	}

	log.WithField("filename", filename).WithField("size", len(contents)).Info("Retrieved file")

	return contents, nil
}

func (s memcacheStore) Delete(filename string) error {
	log.WithField("filename", filename).Debug("Deleting file")

	metadataKey := buildKey(filename)

	metadata, err := s.getKey(metadataKey)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			// File not found. While we may just return nil here, returning an error is more explicit and
			// can surface hidden issues in the code which uses the library
			return ErrFileNotFound
		}

		return fmt.Errorf("Unable to delete file: %w", err)
	}

	totalChunks, _ := strconv.Atoi(string(metadata)) // skipping error checking, will trust metadata not to have anything funny

	err = s.purgeFile(filename, totalChunks)
	if err != nil {
		return fmt.Errorf("Unable to delete file: %w", err)
	}

	log.WithField("filename", filename).Info("Deleted file")

	return nil
}

func (s memcacheStore) purgeFile(filename string, totalChunks int) error {
	for i := 0; i < totalChunks; i++ {
		chunkKey := buildChunkKey(filename, i)
		err := s.deleteKey(chunkKey)
		if err != nil {
			return err
		}
	}

	// Once all chunks deleted, delete the metadata key
	metadataKey := buildKey(filename)
	return s.deleteKey(metadataKey)
}

func (s memcacheStore) setKey(key string, value []byte) error {
	log.WithField("key", key).WithField("size", len(value)).Debug("Setting key")

	return s.client.Set(&memcache.Item{Key: key, Value: value})
}

func (s memcacheStore) getKey(key string) ([]byte, error) {
	log.WithField("key", key).Debug("Getting key")

	item, err := s.client.Get(key)

	if err != nil {
		return []byte{}, err
	}

	log.WithField("key", key).WithField("size", len(item.Value)).Debug("Got key")

	return item.Value, nil
}

func (s memcacheStore) getKeys(keys []string) (map[string][]byte, error) {
	log.WithField("keys", keys).Debug("Getting keys")

	items, err := s.client.GetMulti(keys)
	if err != nil {
		return map[string][]byte{}, err
	}

	if len(items) < len(keys) {
		log.WithField("returned", len(items)).
			WithField("expected", len(keys)).
			Warning("Value count mismatch")

		return map[string][]byte{}, errKeysMissing
	}

	values := map[string][]byte{}
	for _, item := range items {
		log.WithField("key", item.Key).WithField("size", len(item.Value)).Debug("Got key")
		values[item.Key] = item.Value
	}

	return values, nil
}

func (s memcacheStore) deleteKey(key string) error {
	log.WithField("key", key).Debug("Deleting key")

	err := s.client.Delete(key)
	if err == memcache.ErrCacheMiss {
		// The key didnt exist, Memcache must have purged it
		log.WithField("key", key).
			WithError(err).
			Warning("Tried to delete a non existing key")
		return nil
	}

	return err
}

func buildKey(key string) string {
	// Memcache keys cant be longer than 250 chars so lets MD5 variable part of the key (filename)
	// For better performance we can use base64 and md5 only if base64 is longer than 250
	// (or something better than md5 to reduce a chance of collisions)
	hash := md5.Sum([]byte(key))
	hs := hex.EncodeToString(hash[:])
	return keyPrefix + hs
}

func buildChunkKey(key string, index int) string {
	// Use filename + index to identify a specific chunk
	return buildKey(key) + "::" + strconv.Itoa(index)
}
