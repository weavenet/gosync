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
	app.Usage = "gosync OPTIONS SOURCE TARGET"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.IntFlag{"concurrent, c", 20, "number of concurrent transfers"},
		cli.StringFlag{"log-level, l", "info", "log level"},
	}

	const concurrent = 20

	defer log.Flush()

	app.Action = func(c *cli.Context) {
		setLogLevel("info")

		err := validateArgs(c)
		exitOnError(err)

		source := c.Args()[0]
		target := c.Args()[1]

		auth, err := aws.EnvAuth()
		exitOnError(err)
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
	}
	app.Run(os.Args)
}

func validateArgs(c *cli.Context) error {
	if len(c.Args()) != 2 {
		return fmt.Errorf("S3 URL and local directory required.")
	}
	return nil
}

func exitOnError(e error) {
	if e != nil {
		log.Errorf("Received error '%s'", e.Error())
		log.Flush()
		os.Exit(1)
	}
}

func setLogLevel(level string) {
	if level != "info" {
		log.Infof("Setting log level '%s'.", level)
	}
	logConfig := fmt.Sprintf("<seelog minlevel='%s'>", level)
	logger, _ := log.LoggerFromConfigAsBytes([]byte(logConfig))
	log.ReplaceLogger(logger)
}
