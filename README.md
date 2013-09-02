[![Build Status](https://secure.travis-ci.org/brettweavnet/gosync.png)](http://travis-ci.org/brettweavnet/gosync)

# gosync

Sync files, fast.

Gosync leverages go routines to concurrently sync files from S3 to the local file system and vice versa.

# Installation

Clone the repo:

    git clone https://github.com/brettweavnet/gosync

Change into the gosync directory, download the dependencies and run make:

    cd gosync
    go get -d -v ./...
    make

# Setup

Set environment variables:

    AWS_SECRET_ACCESS_KEY=yyy
    AWS_ACCESS_KEY_ID=xxx

# Usage

    gosync sync source target

## Syncing from local directory to S3

    gosync sync /files s3://bucket/files

## Syncing from S3 to local directory

    gosync sync s3://bucket/files /files
