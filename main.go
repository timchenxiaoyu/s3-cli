package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
)

type CmdHandler func(*Config, *cli.Context) error

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "config, c",
			Value: &cli.StringSlice{"$HOME/.s3cfg"},
			Usage: "Config file name.",
		},
	}

	wrapper := func(handler CmdHandler) func(*cli.Context) error {
		return func(c *cli.Context) error {
			config, err := NewConfig(c)
			if err != nil {
				fmt.Println(err)
				return err
			}
			if err := handler(config, c); err != nil {
				fmt.Println(err)
				return err
			}
			return err
		}
	}

	app.Commands = []cli.Command{

		{
			Name:   "ls",
			Usage:  "List objects or buckets -- s3-cli ls [s3://BUCKET[/PREFIX]]",
			Action: wrapper(ListBucket),
			Flags:  app.Flags,
		},
	}

	app.Run(os.Args)
}
