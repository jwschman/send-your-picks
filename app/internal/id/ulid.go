package id

import (
	"math/rand"
	"sync"
	"time"
	//https://github.com/oklog/ulid?tab=readme-ov-file
	"github.com/oklog/ulid/v2"
)

// initialize variables on module initialization
var (
	entropy = ulid.Monotonic(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		0,
	)
	mu sync.Mutex
)

// returns a new ULID
func New() (string, error) {
	mu.Lock()
	defer mu.Unlock()

	id, err := ulid.New(ulid.Timestamp(time.Now()), entropy)
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
