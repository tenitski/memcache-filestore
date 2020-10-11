package handler

import (
	"errors"
	"filestore"
	"net/http"

	"github.com/bouk/httprouter"
	log "github.com/sirupsen/logrus"
)

func NewRetrieveFileHandler(store filestore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithField("request", *r).Debugf("Processing request %s %s", r.Method, r.URL.Path)

		filename := httprouter.GetParam(r, "filename")

		contents, err := store.Retrieve(filename)
		if err != nil {
			log.WithError(err).Error("Error while processing request")

			if errors.Is(err, filestore.ErrFileNotFound) {
				respondWithError(w, r, http.StatusNotFound, err)
				return
			}

			respondWithError(w, r, http.StatusInternalServerError, err)
			return
		}

		respondWithData(w, r, http.StatusOK, contents)
	}
}
