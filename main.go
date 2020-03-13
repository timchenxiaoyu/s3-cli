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
			Action: wrapper(List),
			Flags:  app.Flags,
		},
		{
			Name:   "put",
			Usage:  "Put file into bucket -- s3-cli put FILE [FILE...] s3://BUCKET[/PREFIX]",
			Action: wrapper(Put),
			Flags:  app.Flags,
		},
		{
			Name:   "get",
			Usage:  "Get file from bucket -- s3-cli  get s3://BUCKET/OBJECT LOCAL_FILE",
			Action: wrapper(Get),
			Flags:  app.Flags,
		},
		{
			Name:   "sync",
			Usage:  " Synchronize a directory tree with S3 -- s3-cli  sync LOCAL_DIR s3://BUCKET[/PREFIX] or or s3://BUCKET[/PREFIX] LOCAL_DIR",
			Action: wrapper(Sync),
			Flags:  app.Flags,
		},
		{
			Name:   "rm",
			Usage:  "Delete file from bucket (alias for del) -- s3-cli  rm s3://BUCKET/OBJECT",
			Action: wrapper(Del),
			Flags:  app.Flags,
		},
	}

	app.Run(os.Args)
}
