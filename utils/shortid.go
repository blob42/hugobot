package utils

import (
	"github.com/teris-io/shortid"
)

const (
	Seed = 322124
)

var (
	sid *shortid.Shortid
)

func GetSIDGenerator() *shortid.Shortid {
	return sid
}

func init() {
	var err error
	sid, err = shortid.New(1, shortid.DefaultABC, Seed)
	if err != nil {
		panic(err)
	}
}
