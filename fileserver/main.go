package main

import (
	"filestore"
	"net/http"
	"os"
	"time"

	"github.com/bouk/httprouter"
	log "github.com/sirupsen/logrus"

	"fileserver/handler"
)

func main() {
	// Get server details
	var server string
	if len(os.Args) > 1 {
		server = os.Args[1]
	}
	if server == "" {
		log.Fatalln("Server not provided")
	}

	// Allow to specify log level via ENV var
	if ll := os.Getenv("LOG_LEVEL"); ll != "" {
		logLevel, err := log.ParseLevel(ll)
		if err != nil {
			log.WithError(err).Fatal("Unable to parse log level")
		}
		log.SetLevel(logLevel)
	}

	// Create filestore client
	store := filestore.NewMemcache(server, filestore.MemcacheConfig{
		// Things are not super fast when reading 50MB file, give it plenty of time
		Timeout: 5 * time.Second,

		// For some reason Memcache didnt like 1048576 byte values in my setup, the max value it would take is 1048470...
		// Memcache logs show that it gets exactly specified number of bytes, no envelop seem to be added by
		// third party Memcache client lib.
		ChunkSize: 1048470,
	})

	// Configure routes
	router := httprouter.New()
	router.POST("/file/:filename", handler.NewStoreFileHandler(store))
	router.GET("/file/:filename", handler.NewRetrieveFileHandler(store))
	router.DELETE("/file/:filename", handler.NewDeleteFileHandler(store))

	// Start server
	addr := ":8080"
	log.WithField("addr", addr).Info("Starting server")
	log.Fatal(http.ListenAndServe(addr, router))
}
