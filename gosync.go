package main

import (
	"fmt"
	"os"

	"github.com/brettweavnet/gosync/gosync"
	log "github.com/cihub/seelog"
	"github.com/codegangsta/cli"
	"github.com/mitchellh/goamz/aws"
)

func main() {
	app := cli.NewApp()
	app.Name = "gosync"
	app.Usage = "Concurrently sync files to/from S3."
	app.Version = "0.0.1"

	const concurrent = 20

	defer log.Flush()

	app.Commands = []cli.Command{
		{
			Name:        "sync",
			Usage:       "gosync sync SOURCE TARGET",
			Description: "Sync directories to / from S3 bucket.",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 2 {
					log.Errorf("S3 URL and local directory required.")
					os.Exit(1)
				}
				source := c.Args()[0]
				target := c.Args()[1]
				setLogLevel("info")

				auth, err := aws.EnvAuth()
				if err != nil {
					log.Errorf("Error loading AWS credentials: %s", err)
					os.Exit(1)
				}

				log.Infof("Syncing '%s' with '%s'", source, target)

				sync := gosync.SyncPair{source, target, auth, concurrent}
				err = sync.Sync()
				if err == nil {
					log.Infof("Syncing completed succesfully.")
				} else {
					fmt.Printf("Sync failed: %s", err)
					os.Exit(1)
				}
			},
		},
	}
	app.Run(os.Args)
}

func setLogLevel(level string) {
	if level != "info" {
		log.Infof("Setting log level '%s'.", level)
	}
	logConfig := fmt.Sprintf("<seelog minlevel='%s'>", level)
	logger, _ := log.LoggerFromConfigAsBytes([]byte(logConfig))
	log.ReplaceLogger(logger)
}
