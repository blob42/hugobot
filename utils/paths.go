package utils

import (
	"os"
	"path/filepath"
)

const (
	MkdirMode = 0770
)

func Mkdir(path ...string) error {
	joined := filepath.Join(path...)

	return os.MkdirAll(joined, MkdirMode)
}
