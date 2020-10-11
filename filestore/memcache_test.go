package filestore

import (
	"filestore/client"
	"filestore/mock"
	"fmt"
	"reflect"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
)

// A bunch of test below, other potential tests may include:
// - Are there correct number of chunks for each written file (currently it only checks for know chunks)
// - Test storing zero length file
// - Test storing file over max size
// - Test storing file which already exists
// - Test retrieval of existing, non existing, corrupted files
// - Test deletion of existing, non existing, corrupted files

func TestHandler_Store(t *testing.T) {
	type testEnv struct {
		client client.Memcache
		config MemcacheConfig
	}

	type args struct {
		filename string
		contents []byte
	}

	type testCase struct {
		name      string
		env       testEnv
		args      args
		wantKeys  map[string][]byte
		wantError error
	}

	c := mock.NewMemcacheClient(4) // allow to save up to 4 keys

	defaultEnv := testEnv{
		client: c,
		config: MemcacheConfig{
			ChunkSize:   10,
			MaxFileSize: 500,
		},
	}

	tests := []testCase{
		func() testCase {
			filename := "file.dat"
			contents := []byte("some content")

			return testCase{
				name: "Successfully saved a file",
				env:  defaultEnv,
				args: args{filename, contents},
				wantKeys: map[string][]byte{
					buildKey(filename):         []byte("2"),
					buildChunkKey(filename, 0): []byte("some conte"),
					buildChunkKey(filename, 1): []byte("nt"),
				},
				wantError: nil,
			}
		}(),

		func() testCase {
			filename := "file.dat"
			contents := []byte("some content")

			return testCase{
				name:      "Failed to save a file as it already exists",
				env:       defaultEnv,
				args:      args{filename, contents},
				wantKeys:  map[string][]byte{},
				wantError: ErrFileAlreadyExists,
			}
		}(),

		func() testCase {
			filename := "large-file.dat"
			contents := []byte("some too large content")

			return testCase{
				name:      "Failed to fit file to the storage, chunks lost",
				env:       defaultEnv,
				args:      args{filename, contents},
				wantKeys:  map[string][]byte{},
				wantError: fmt.Errorf("Unable to store file: %w", ErrFileCorrupted),
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewMemcacheWithClient(tt.env.client, tt.env.config).Store(tt.args.filename, tt.args.contents)

			for key, wantVal := range tt.wantKeys {
				var val []byte
				item, err := tt.env.client.Get(key)
				if err != nil && err != memcache.ErrCacheMiss {
					panic(err)
				} else {
					val = item.Value
				}

				if !reflect.DeepEqual(val, wantVal) {
					t.Errorf("Key %s: want %#v, got %#v", key, wantVal, val)
				}
			}

			if !reflect.DeepEqual(err, tt.wantError) {
				t.Errorf("Error: want %#v, got %#v", tt.wantError, err)
			}
		})
	}
}
