package handler

import (
	"context"
	"fileserver/mock"
	"filestore"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bouk/httprouter"
)

var notFoundResponse = []byte(`{
  "error": "File not found"
}`)

func TestHandler_DeleteFileHandler(t *testing.T) {
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
			filename := "non-existing-file.dat"

			ctx := httprouter.WithParams(context.Background(), httprouter.Params{httprouter.Param{
				Key:   "filename",
				Value: filename,
			}})

			request := httptest.NewRequest(http.MethodDelete, "/files/"+filename, nil).WithContext(ctx)

			return testCase{
				name:       "Deleting a non existing file",
				env:        defaultEnv,
				args:       args{request},
				wantCode:   http.StatusNotFound,
				wantBody:   notFoundResponse,
				wantHeader: http.Header{"Content-Type": []string{"application/json"}},
			}
		}(),

		func() testCase {
			filename := "existing-file.dat"

			ctx := httprouter.WithParams(context.Background(), httprouter.Params{httprouter.Param{
				Key:   "filename",
				Value: filename,
			}})

			request := httptest.NewRequest(http.MethodDelete, "/files/"+filename, nil).WithContext(ctx)

			return testCase{
				name:       "Deleting an existing file",
				env:        defaultEnv,
				args:       args{request},
				wantCode:   http.StatusOK,
				wantBody:   nil,
				wantHeader: http.Header{},
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			NewDeleteFileHandler(tt.env.store).ServeHTTP(recorder, tt.args.request)

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
