package main

import (
	"fmt"
	"github.com/brettweavnet/gosync/gosync"
	"github.com/codegangsta/cli"
	"launchpad.net/goamz/aws"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "gosync"
	app.Usage = "CLI for S3"

	const concurrent = 20

	app.Commands = []cli.Command{
		{
			Name:        "sync",
			Usage:       "gosync sync SOURCE TARGET",
			Description: "Sync directories to / from S3 bucket.",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 2 {
					fmt.Printf("S3 URL and local directory required.")
					os.Exit(1)
				}
				source := c.Args()[0]
				target := c.Args()[1]
				auth, err := aws.EnvAuth()
				if err != nil {
					fmt.Printf("Error loading AWS credentials: %s", err)
					os.Exit(1)
				}

				fmt.Printf("Syncing %s with %s\n", source, target)

				sync := gosync.SyncPair{source, target, auth, concurrent}
				err = sync.Sync()
				if err == nil {
					fmt.Printf("Syncing completed succesfully.")
				} else {
					fmt.Printf("Sync failed: %s", err)
					os.Exit(1)
				}
			},
		},
	}
	app.Run(os.Args)
}
