package handler

import (
	"errors"
	"filestore"
	"io/ioutil"
	"net/http"

	"github.com/bouk/httprouter"

	log "github.com/sirupsen/logrus"
)

func NewStoreFileHandler(store filestore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithField("request", *r).Debugf("Processing request %s %s", r.Method, r.URL.Path)

		filename := httprouter.GetParam(r, "filename")

		contents, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.WithError(err).Error("Error while reading request body")
			respondWithStatusCode(w, r, http.StatusBadRequest)
			return
		}

		err = store.Store(filename, contents)
		if err != nil {
			log.WithError(err).Error("Error while processing request")

			if errors.Is(err, filestore.ErrFileAlreadyExists) {
				respondWithError(w, r, http.StatusConflict, err)
				return
			}

			if errors.Is(err, filestore.ErrFileTooLarge) {
				respondWithError(w, r, http.StatusBadRequest, err)
				return
			}

			respondWithError(w, r, http.StatusInternalServerError, err)
			return
		}

		respondWithStatusCode(w, r, http.StatusOK)
	}
}
