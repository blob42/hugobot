package export

import (
	"hugobot/bitcoin"
	"hugobot/config"
	"hugobot/encoder"
	"log"
	"os"
	"path/filepath"

	qrcode "github.com/skip2/go-qrcode"
)

func ExportBTCAddresses() error {
	unusedAddrs, err := bitcoin.GetAllUsedUnsyncedAddresses()
	if err != nil {
		return err
	}

	for _, a := range unusedAddrs {
		//first export the qr codes
		log.Println("exporting ", a)

		qrFileName := a.Address + ".png"

		qrCodePath := filepath.Join(config.RelBitcoinAddrContentPath(),
			config.BTCQRCodesDir, qrFileName)

		err := qrcode.WriteFile(a.Address, qrcode.Medium, 580, qrCodePath)
		if err != nil {
			return err
		}

		// store the address pages

		filename := a.Address + ".md"
		filePath := filepath.Join(config.RelBitcoinAddrContentPath(), filename)

		data := map[string]interface{}{
			"linked_article_id": a.LinkedArticleID,
			//"resources": []map[string]interface{}{
			//map[string]interface{}{
			//"src": filepath.Join(config.BTCQRCodesDir, a.Address+".png"),
			//},
			//},
		}

		addressPage, err := os.Create(filePath)
		if err != nil {
			return err
		}

		tomlExporter := encoder.NewExportEncoder(addressPage, encoder.TOML)
		tomlExporter.Encode(data)

		// Set synced
		err = a.SetSynced()
		if err != nil {
			return err
		}

	}

	return nil
}
