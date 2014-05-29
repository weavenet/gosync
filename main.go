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

	app.Action = func(c *cli.Context) {
		defer log.Flush()
		setLogLevel(c.String("log-level"))

		err := validateArgs(c)
		exitOnError(err)

		log.Debugf("Reading AWS credentials from AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.")
		auth, err := aws.EnvAuth()
		exitOnError(err)

		source := c.Args()[0]
		log.Infof("Setting source to '%s'.", source)

		target := c.Args()[1]
		log.Infof("Setting target to '%s'.", target)

		sync := gosync.NewSync(auth, source, target)

		sync.Concurrent = c.Int("concurrent")
		log.Infof("Setting concurrent transfers to '%d'.", sync.Concurrent)

		err = sync.Sync()
		exitOnError(err)

		log.Infof("Syncing completed successfully.")
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
