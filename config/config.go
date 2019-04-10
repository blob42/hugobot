package config

import (
	"log"
	"path"

	"github.com/fatih/structs"
)

const (
	BTCQRCodesDir = "qrcodes"
)

type Config struct {
	WebsitePath               string
	GithubAccessToken         string
	RelBitcoinAddrContentPath string
	ApiPort                   int
}

var (
	C *Config
)

func HugoData() string {
	return path.Join(C.WebsitePath, "data")
}

func HugoContent() string {
	return path.Join(C.WebsitePath, "content")
}

func RelBitcoinAddrContentPath() string {
	return path.Join(C.WebsitePath, C.RelBitcoinAddrContentPath)
}

func RegisterConf(conf string, val interface{}) error {
	log.Printf("Setting %#v to %#v", conf, val)
	s := structs.New(C)

	field, ok := s.FieldOk(conf)

	// Conf option not registered in Config struct
	if !ok {
		return nil
	}

	err := field.Set(val)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	C = new(Config)
}
