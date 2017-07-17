package gosync

import (
	"io/ioutil"
	"mime"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/goamz/s3"
)

func (s *SyncPair) syncDirToS3() error {
	log.Infof("Syncing to S3.")

	sourceFiles, err := loadLocalFiles(s.Source)
	if err != nil {
		return err
	}

	s3url := newS3Url(s.Target)
	path := s3url.Path()

	bucket, err := lookupBucket(s3url.Bucket(), s.Auth, s.Region)
	if err != nil {
		return err
	}

	// Load files and do not specify marker to start
	targetFiles := make(map[string]string)
	targetFiles, err = loadS3Files(bucket, path, targetFiles, "")
	if err != nil {
		return err
	}

	return s.concurrentSyncDirToS3(s3url, bucket, targetFiles, sourceFiles)
}

func (s *SyncPair) concurrentSyncDirToS3(s3url s3Url, bucket *s3.Bucket, targetFiles, sourceFiles map[string]string) error {
	doneChan := newDoneChan(s.Concurrent)
	pool := newPool(s.Concurrent)
	var wg sync.WaitGroup

	for file, _ := range sourceFiles {
		// ensure the file has no leading slashes to it compares correctly
		relativeTargetFile := strings.TrimLeft(strings.Join([]string{s3url.Path(), file}, "/"), "/")

		if targetFiles[relativeTargetFile] != sourceFiles[file] {
			filePath := strings.Join([]string{s.Source, file}, "/")
			keyPath := strings.Join([]string{s3url.Key(), file}, "/")

			// Get transfer reservation from pool
			log.Tracef("Requesting reservation for '%s'.", keyPath)
			<-pool
			log.Tracef("Retrieved reservation for '%s'.", keyPath)

			log.Infof("Starting sync: %s -> s3://%s/%s", filePath, bucket.Name, file)
			wg.Add(1)
			go func(doneChan chan error, filePath string, bucket *s3.Bucket, keyPath string) {
				defer wg.Done()
				writeLocalFileToS3Routine(doneChan, filePath, bucket, keyPath)
				pool <- 1
			}(doneChan, filePath, bucket, keyPath)
		}
	}

	// Wait for all routines to finish
	wg.Wait()
	return nil
}

func writeLocalFileToS3Routine(doneChan chan error, filePath string, bucket *s3.Bucket, file string) {
	err := writeLocalFileToS3(bucket, file, filePath)
	if err != nil {
		doneChan <- err
	}
	log.Infof("Sync completed successfully: %s -> s3://%s/%s.", filePath, bucket.Name, file)
	doneChan <- nil
}

func writeLocalFileToS3(bucket *s3.Bucket, path string, file string) error {
	contType := mime.TypeByExtension(filepath.Ext(file))
	Perms := s3.ACL("private")

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	if err := bucket.Put(path, data, contType, Perms); err != nil {
		return err
	}

	return nil
}
