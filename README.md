[![Build Status](https://secure.travis-ci.org/brettweavnet/gosync.png)](http://travis-ci.org/brettweavnet/gosync)

# gosync

I want to be the fastest way to concurrently sync files and directories to/from S3.

# Installation

Ensure you have Go 1.2 or greater installed and your GOPATH is set.

Clone the repo:

    go get github.com/brettweavnet/gosync

Change into the gosync directory and run make:

    cd $GOPATH/src/github.com/brettweavnet/gosync/
    make

# Setup

Set environment variables:

    AWS_SECRET_ACCESS_KEY=yyy
    AWS_ACCESS_KEY_ID=xxx

# Usage

    gosync sync SOURCE TARGET

## Syncing from local directory to S3

    gosync /files s3://bucket/files

## Syncing from S3 to local directory

    gosync s3://bucket/files /files

# Known Issues

* gosync does not currently support removing files from the target.
* Only supports setting default permissions on local files and private in s3.

# Contributing

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
4. Push to the branch (git push origin my-new-feature)
5. Create new Pull Request
