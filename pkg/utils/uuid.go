package utils

import (
	"crypto/rand"
	"github.com/oklog/ulid/v2"
)

func GenerateConnID() string {
	return ulid.MustNew(ulid.Now(), rand.Reader).String()
}
