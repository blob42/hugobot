package main

import (
	"log"
	"os"

	"git.blob42.xyz/blob42/hugobot/v3/config"
	"git.blob42.xyz/blob42/hugobot/v3/logging"

	"github.com/gobuffalo/flect"
	altsrc "github.com/urfave/cli/altsrc"
	cli "gopkg.in/urfave/cli.v1"
)

func tearDown() {
	DB.Handle.Close()
	logging.Close()
}

func main() {
	defer tearDown()

	app := cli.NewApp()
	app.Name = "hugobot"
	app.Version = "1.0"
	flags := []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "website-path",
			Usage:  "`PATH` to hugo project",
			EnvVar: "WEBSITE_PATH",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "github-access-token",
			Usage:  "Github API Access Token",
			EnvVar: "GH_ACCESS_TOKEN",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "rel-bitcoin-addr-content-path",
			Usage: "path to bitcoin data relative to hugo path",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "api-port",
			Usage: "default bot api port",
		}),

		cli.StringFlag{
			Name:  "config",
			Value: "config.toml",
			Usage: "TOML config `FILE` path",
		},
	}

	app.Before = func(c *cli.Context) error {

		err := altsrc.InitInputSourceWithContext(flags,
			altsrc.NewTomlSourceFromFlagFunc("config"))(c)
		if err != nil {
			return err
		}

		for _, conf := range c.GlobalFlagNames() {

			// find corresponding flag
			for _, flag := range flags {
				if flag.GetName() == conf {
					switch flag.(type) {
					case cli.StringFlag:
						err = config.RegisterConf(flect.Pascalize(conf), c.GlobalString(conf))
					case *altsrc.StringFlag:
						err = config.RegisterConf(flect.Pascalize(conf), c.GlobalString(conf))
					case cli.IntFlag:
						err = config.RegisterConf(flect.Pascalize(conf), c.GlobalInt(conf))
					case *altsrc.IntFlag:
						err = config.RegisterConf(flect.Pascalize(conf), c.GlobalInt(conf))

					}
				}
			}

		}

		return err
	}

	app.Flags = flags

	app.Commands = []cli.Command{
		startServerCmd,
		exportCmdGrp,
		feedsCmdGroup,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
