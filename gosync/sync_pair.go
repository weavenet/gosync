package gosync

import (
	"errors"
	"strings"

	"github.com/mitchellh/goamz/aws"
)

type SyncPair struct {
	Auth       aws.Auth
	Source     string
	Target     string
	Concurrent int
	Region string
}

func NewSyncPair(auth aws.Auth, source string, target string, region string) *SyncPair {
	return &SyncPair{
		Auth:       auth,
		Source:     source,
		Target:     target,
		Concurrent: 1,
		Region: region,
	}
}

func (s *SyncPair) Sync() error {
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

func (s *SyncPair) validPair() bool {
	if !validS3Url(s.Source) && !validS3Url(s.Target) {
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
