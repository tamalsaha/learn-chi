package main

import (
	"fmt"
	"github.com/oklog/ulid/v2"
	"io"
	"math/rand"
	"sync"
	"time"
)

var pool = sync.Pool{
	New: func() interface{} {
		return ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	},
}

func New() (ulid.ULID, error) {
	entropy := pool.Get()
	result, err := ulid.New(ulid.Timestamp(time.Now()), entropy.(io.Reader))
	pool.Put(entropy)
	return result, err
}

func MustNew() ulid.ULID {
	entropy := pool.Get()
	defer pool.Put(entropy)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy.(io.Reader))
}

func main() {
	fmt.Println(New())
	fmt.Println(MustNew())
	// Output: 0000XSNJG0MQJHBF4QX1EFD6Y3
}
