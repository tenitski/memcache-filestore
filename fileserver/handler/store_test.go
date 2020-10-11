package handler

import (
	"context"
	"fileserver/mock"
	"filestore"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/bouk/httprouter"
)

var fileAlreadyExistsResponse = []byte(`{
  "error": "File already exists"
}`)

var fileTooLargeResponse = []byte(`{
  "error": "File is too large: max file size is 50 bytes"
}`)

func TestHandler_StoreFileHandler(t *testing.T) {
	type testEnv struct {
		store filestore.Store
	}

	type args struct {
		request *http.Request
	}

	type testCase struct {
		name       string
		env        testEnv
		args       args
		wantCode   int
		wantBody   []byte
		wantHeader http.Header
	}

	store := mock.NewFilestore(50)
	err := store.Store("existing-file.dat", []byte("some contents"))
	if err != nil {
		panic(err)
	}

	defaultEnv := testEnv{
		store: store,
	}

	tests := []testCase{
		func() testCase {
			filename := "new-file.dat"
			contents := []byte("some content")

			ctx := httprouter.WithParams(context.Background(), httprouter.Params{httprouter.Param{
				Key:   "filename",
				Value: filename,
			}})

			request := httptest.NewRequest(http.MethodPost, "/files/"+filename, strings.NewReader(string(contents))).WithContext(ctx)

			return testCase{
				name:       "Storing a new file",
				env:        defaultEnv,
				args:       args{request},
				wantCode:   http.StatusOK,
				wantBody:   nil,
				wantHeader: http.Header{},
			}
		}(),

		func() testCase {
			filename := "existing-file.dat"
			contents := []byte("some content")

			ctx := httprouter.WithParams(context.Background(), httprouter.Params{httprouter.Param{
				Key:   "filename",
				Value: filename,
			}})

			request := httptest.NewRequest(http.MethodPost, "/files/"+filename, strings.NewReader(string(contents))).WithContext(ctx)

			return testCase{
				name:       "Storing an existing file",
				env:        defaultEnv,
				args:       args{request},
				wantCode:   http.StatusConflict,
				wantBody:   fileAlreadyExistsResponse,
				wantHeader: http.Header{"Content-Type": []string{"application/json"}},
			}
		}(),

		func() testCase {
			filename := "new-large-file.dat"
			contents := []byte("some very very very very very very very very very very very large content")

			ctx := httprouter.WithParams(context.Background(), httprouter.Params{httprouter.Param{
				Key:   "filename",
				Value: filename,
			}})

			request := httptest.NewRequest(http.MethodPost, "/files/"+filename, strings.NewReader(string(contents))).WithContext(ctx)

			return testCase{
				name:       "Storing a new too large file",
				env:        defaultEnv,
				args:       args{request},
				wantCode:   http.StatusBadRequest,
				wantBody:   fileTooLargeResponse,
				wantHeader: http.Header{"Content-Type": []string{"application/json"}},
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			NewStoreFileHandler(tt.env.store).ServeHTTP(recorder, tt.args.request)

			if !reflect.DeepEqual(recorder.Code, tt.wantCode) {
				t.Errorf("Code: want %#v, got %#v", tt.wantCode, recorder.Code)
			}

			if !reflect.DeepEqual(recorder.Body.Bytes(), tt.wantBody) {
				t.Errorf("Body: want %#v, got %#v", string(tt.wantBody), recorder.Body.String())
			}

			if !reflect.DeepEqual(recorder.Header(), tt.wantHeader) {
				t.Errorf("Header: want %#v, got %#v", tt.wantHeader, recorder.Header())
			}
		})
	}
}
