package gosync

import (
	"errors"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/goamz/aws"
)

type Sync struct {
	Auth       aws.Auth
	Source     string
	Target     string
	Concurrent int
}

func NewSync(auth aws.Auth, source string, target string) *Sync {
	return &Sync{
		Auth:       auth,
		Source:     source,
		Target:     target,
		Concurrent: 1,
	}
}

func (s *Sync) Sync() error {
	if !s.validPair() {
		return errors.New("Invalid sync pair.")
	}

	if validS3Url(s.Source) && validS3Url(s.Target) {
		return s.syncS3ToS3()
	}

	if validS3Url(s.Source) {
		return s.syncS3ToDir()
	}

	return s.syncDirToS3()
}

func (s *Sync) validPair() bool {
	if !validS3Url(s.Source) || !validS3Url(s.Target) {
		return false
	}

	if validTarget(s.Source) && validTarget(s.Target) {
		return true
	}
	return false
}

func validTarget(target string) bool {
	// Check for local file
	if pathExists(target) {
		return true
	}

	// Check for valid s3 url
	if validS3Url(target) {
		return true
	}

	return false
}

func validS3Url(path string) bool {
	return strings.HasPrefix(path, "s3://")
}

func waitForRoutines(routines []chan string) {
	for _, r := range routines {
		msg := <-r
		log.Infof("%s", msg)
	}
}

func relativePath(path string, filePath string) string {
	if path == "." {
		return strings.TrimPrefix(filePath, "/")
	} else {
		return strings.TrimPrefix(strings.TrimPrefix(filePath, path), "/")
	}
}

func newDoneChan(concurrent int) chan error {
	// Panic on any errors
	doneChan := make(chan error, concurrent)
	go func() {
		for {
			select {
			case err := <-doneChan:
				if err != nil {
					panic(err.Error())
				}
			}
		}
	}()
	return doneChan
}
