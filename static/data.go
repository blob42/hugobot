package static

import (
	"git.blob42.xyz/blob42/hugobot/v3/config"
	"encoding/json"
	"os"
	"path/filepath"
)

var data = map[string]interface{}{
	"bolts": map[string]interface{}{
		"names": BoltNames,
	},
}

// Json Export Static Data
func HugoExportData() error {
	dirPath := filepath.Join(config.HugoData())
	for k, v := range data {
		filePath := filepath.Join(dirPath, k+".json")
		outputFile, err := os.Create(filePath)
		defer outputFile.Close()
		if err != nil {
			return err
		}

		jsonEnc := json.NewEncoder(outputFile)
		jsonEnc.Encode(v)
	}

	return nil
}
