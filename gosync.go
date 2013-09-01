package main

import (
    "os"
    "fmt"
    "github.com/codegangsta/cli"
    "github.com/brettweavnet/gosync/gosync"
    "launchpad.net/goamz/aws"
)

func main() {
    app := cli.NewApp()
    app.Name = "gosync"
    app.Usage = "CLI for S3"

    app.Commands = []cli.Command{
      {
        Name:        "sync",
        Usage:       "gosync sync LOCAL_DIR s3://BUCKET/KEY",
        Description: "Sync local dir with S3 URL.",
        Action: func(c *cli.Context) {
          if len(c.Args()) < 2 {
             fmt.Printf("S3 URL and local directory required.")
             os.Exit(1)
          }
          arg0 := c.Args()[0]
          arg1 := c.Args()[1]
          auth, err := aws.EnvAuth()
          if err != nil {
              panic(err)
          }

          fmt.Printf("Syncing %s with %s\n", arg0, arg1)

          sync := gosync.SyncPair{arg0, arg1, auth}
          result := sync.Sync()
          if result == true {
              fmt.Printf("Syncing completed succesfully.")
          } else {
              fmt.Printf("Syncing failed.")
              os.Exit(1)
          }
        },
      },
    }
    app.Run(os.Args)
}
