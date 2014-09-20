package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/brettweavnet/gosync/gosync"
	"github.com/brettweavnet/gosync/version"

	log "github.com/cihub/seelog"
	"github.com/codegangsta/cli"
	"github.com/mitchellh/goamz/aws"
)

func main() {
	app := cli.NewApp()
	app.Name = "gosync"
	app.Usage = "gosync OPTIONS SOURCE TARGET"
	app.Version = version.Version()
	app.Flags = []cli.Flag{
		cli.IntFlag{"concurrent, c", 20, "number of concurrent transfers", ""},
		cli.StringFlag{"log-level, l", "info", "log level", ""},
		cli.StringFlag{"aws-secret-access-key", "", "AWS Secret Access Key", ""},
		cli.StringFlag{"aws-access-key-id", "", "AWS Access Key Id", ""},
		cli.StringFlag{"aws-security-token", "", "AWS Security Token", ""},
	}

	const concurrent = 20

	app.Action = func(c *cli.Context) {
		defer log.Flush()
		setLogLevel(c.String("log-level"))

		err := validateArgs(c)
		exitOnError(err)

		key := strings.TrimSpace(c.String("aws-access-key-id"))
		secret := strings.TrimSpace(c.String("aws-secret-access-key"))
		token := strings.TrimSpace(c.String("aws-security-token"))

		auth, err := aws.GetAuth(key, secret)
		exitOnError(err)
		if token != "" {
			auth.Token = token
		}

		source := c.Args()[0]
		log.Infof("Setting source to '%s'.", source)

		target := c.Args()[1]
		log.Infof("Setting target to '%s'.", target)

		syncPair := gosync.NewSyncPair(auth, source, target)

		syncPair.Concurrent = c.Int("concurrent")
		log.Infof("Setting concurrent transfers to '%d'.", syncPair.Concurrent)

		err = syncPair.Sync()
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
	if level != "error" && level != "warn" {
		log.Infof("Setting log level '%s'.", level)
	}
	logConfig := fmt.Sprintf("<seelog minlevel='%s'>", level)
	logger, _ := log.LoggerFromConfigAsBytes([]byte(logConfig))
	log.ReplaceLogger(logger)
}
