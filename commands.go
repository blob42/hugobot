package main

import (
	"git.blob42.xyz/blob42/hugobot/v3/export"
	"git.blob42.xyz/blob42/hugobot/v3/feeds"
	"git.blob42.xyz/blob42/hugobot/v3/static"
	"log"

	cli "gopkg.in/urfave/cli.v1"
)

var startServerCmd = cli.Command{
	Name:    "server",
	Aliases: []string{"s"},
	Usage:   "Run server",
	Action:  startServer,
}

var exportCmdGrp = cli.Command{
	Name:    "export",
	Aliases: []string{"e"},
	Usage:   "Export to hugo",
	Subcommands: []cli.Command{
		exportPostsCmd,
		exportWeeksCmd,
		exportBTCAddressesCmd,
	},
}

var exportBTCAddressesCmd = cli.Command{
	Name:   "btc",
	Usage:  "export bitcoin addresses",
	Action: exportAddresses,
}

var exportWeeksCmd = cli.Command{
	Name:   "weeks",
	Usage:  "export weeks",
	Action: exportWeeks,
}

var exportPostsCmd = cli.Command{
	Name:   "posts",
	Usage:  "Export posts to hugo",
	Action: exportPosts,
}

func startServer(c *cli.Context) {
	server()
}

func exportPosts(c *cli.Context) {
	exporter := export.NewHugoExporter()
	feeds, err := feeds.ListFeeds()
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range feeds {
		exporter.Export(*f)
	}

	// Export static data
	err = static.HugoExportData()
	if err != nil {
		log.Fatal(err)
	}

}

func exportWeeks(c *cli.Context) {
	err := export.ExportWeeks()
	if err != nil {
		log.Fatal(err)
	}

}

func exportAddresses(c *cli.Context) {
	err := export.ExportBTCAddresses()
	if err != nil {
		log.Fatal(err)
	}

}
