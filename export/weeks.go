// Export all weeks to the weeks content directory
package export

import (
	"git.blob42.xyz/blob42/hugobot/v3/config"
	"git.blob42.xyz/blob42/hugobot/v3/encoder"
	"git.blob42.xyz/blob42/hugobot/v3/utils"
	"os"
	"path/filepath"
	"time"
)

const (
	FirstWeek = "2017-12-07"
)

var (
	WeeksContentDir = "weeks"
)

type WeekData struct {
	Title string
	Date  time.Time
}

func ExportWeeks() error {
	firstWeek, err := time.Parse("2006-01-02", FirstWeek)
	if err != nil {
		return err
	}

	WeeksTilNow := utils.GetAllThursdays(firstWeek, time.Now())
	for _, week := range WeeksTilNow {
		weekName := week.Format("2006-01-02")
		fileName := weekName + ".md"

		weekFile, err := os.Create(filepath.Join(config.HugoContent(),
			WeeksContentDir,
			fileName))

		if err != nil {
			return err
		}

		weekData := WeekData{
			Title: weekName,
			Date:  week,
		}

		tomlExporter := encoder.NewExportEncoder(weekFile, encoder.TOML)

		tomlExporter.Encode(weekData)
	}

	return nil
}
