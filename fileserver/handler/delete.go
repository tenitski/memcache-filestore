package handler

import (
	"errors"
	"net/http"

	"filestore"

	"github.com/bouk/httprouter"
	log "github.com/sirupsen/logrus"
)

func NewDeleteFileHandler(store filestore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithField("request", *r).Debugf("Processing request %s %s", r.Method, r.URL.Path)

		filename := httprouter.GetParam(r, "filename")

		err := store.Delete(filename)
		if err != nil {
			log.WithError(err).Error("Error while processing request")

			if errors.Is(err, filestore.ErrFileNotFound) {
				respondWithError(w, r, http.StatusNotFound, err)
				return
			}

			respondWithError(w, r, http.StatusInternalServerError, err)
			return
		}

		respondWithStatusCode(w, r, http.StatusOK)
	}
}
