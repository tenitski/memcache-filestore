package main

import (
	"filestore"
	"io/ioutil"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	var server string
	if len(os.Args) > 1 {
		server = os.Args[1]
	}

	if server == "" {
		log.Fatal("Server not provided")
	}

	var filename string
	if len(os.Args) > 2 {
		filename = os.Args[2]
	}

	if filename == "" {
		log.Fatal("Filename not provided")
	}

	log.SetLevel(log.DebugLevel)

	c := filestore.NewMemcache(server, filestore.MemcacheConfig{
		// Things are not super fast when reading 50MB file, give it plenty of time
		Timeout: 5 * time.Second,

		// For some reason Memcache didnt like 1048576 byte values in my setup, the max value it would take is 1048470...
		// Memcache logs show that it gets exactly specified number of bytes, no envelop seem to be added by the used
		// Memcache client lib.
		ChunkSize: 1048470,
	})

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.WithError(err).Fatal("Unable to read original file")
	}

	err = c.Store(filename, contents)
	if err != nil {
		log.WithError(err).Fatal("Unable to store file")
	}

	retrievedContents, err := c.Retrieve(filename)
	if err != nil {
		log.WithError(err).Fatal("Unable to retrieve file")
	}

	err = ioutil.WriteFile(filename+"-retrieved", retrievedContents, 0644)
	if err != nil {
		log.WithError(err).Fatal("Unable to save to file")
	}

	err = c.Delete(filename)
	if err != nil {
		log.WithError(err).Fatal("Unable to delete file")
	}
}
