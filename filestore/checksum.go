package filestore

import (
	"crypto/md5"
	"encoding/hex"
)

func checksum(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
