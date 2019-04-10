package encoder

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
)

const (
	JSON = iota
	TOML
)

type Encoder interface {
	Encode(v interface{}) error
}

type ExportEncoder struct {
	encoder Encoder
	w       io.Writer
	eType   int
}

func (ee *ExportEncoder) Encode(v interface{}) error {
	var err error

	if ee.eType == TOML {
		fmt.Fprintf(ee.w, "+++\n")
	}

	err = ee.encoder.Encode(v)

	if ee.eType == TOML {
		fmt.Fprintf(ee.w, "+++\n")
	}

	return err
}

func NewExportEncoder(w io.Writer, encType int) *ExportEncoder {

	var enc Encoder

	switch encType {
	case JSON:
		enc = json.NewEncoder(w)
	case TOML:
		enc = toml.NewEncoder(w)
	}

	return &ExportEncoder{
		encoder: enc,
		w:       w,
		eType:   encType,
	}
}
