package gosync

import (
	"mime"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/goamz/s3"
)

func (s *Sync) syncS3ToS3() error {
	log.Infof("Syncing from S3 to S3.")

	sourceS3Url := newS3Url(s.Source)
	sourceBucket, err := lookupBucket(sourceS3Url.Bucket(), s.Auth)
	if err != nil {
		return err
	}

	targetS3Url := newS3Url(s.Target)
	targetBucket, err := lookupBucket(targetS3Url.Bucket(), s.Auth)
	if err != nil {
		return err
	}

	return s.concurrentSyncS3ToS3(sourceS3Url, targetS3Url, sourceBucket, targetBucket)
}

func (s *Sync) concurrentSyncS3ToS3(sourceS3Url, targetS3Url s3Url, sourceBucket, targetBucket *s3.Bucket) error {
	doneChan := newDoneChan(s.Concurrent)
	pool := newPool(s.Concurrent)
	var wg sync.WaitGroup

	sourceFiles := make(map[string]string)
	sourceFiles, err := loadS3Files(sourceBucket, sourceS3Url.Path(), sourceFiles, "")
	if err != nil {
		return err
	}

	targetFiles := make(map[string]string)
	targetFiles, err = loadS3Files(targetBucket, targetS3Url.Path(), targetFiles, "")
	if err != nil {
		return err
	}

	for file, _ := range sourceFiles {
		if targetFiles[file] != sourceFiles[file] {
			sourceKeyPath := strings.Join([]string{sourceS3Url.Key(), file}, "/")
			targetKeyPath := strings.Join([]string{targetS3Url.Key(), file}, "/")

			// Get transfer reservation from pool
			log.Tracef("Requesting reservation for '%s'.", file)
			<-pool
			log.Tracef("Retrieved reservation for '%s'.", file)

			log.Infof("Starting sync: s3://%s/%s -> s3://%s/%s.", sourceBucket.Name, sourceKeyPath, targetBucket.Name, targetKeyPath)
			wg.Add(1)
			go func(doneChan chan error, sourceBucket, targetBucket *s3.Bucket, sourceKeyPath, targetKeyPath string) {
				defer wg.Done()
				writeS3FileToS3Routine(doneChan, sourceBucket, targetBucket, sourceKeyPath, targetKeyPath)
				pool <- 1
			}(doneChan, sourceBucket, targetBucket, sourceKeyPath, targetKeyPath)
		}
	}

	wg.Wait()
	return nil
}

func writeS3FileToS3Routine(doneChan chan error, sourceBucket, targetBucket *s3.Bucket, sourceKeyPath, targetKeyPath string) {
	err := writeS3FileToS3(sourceBucket, targetBucket, sourceKeyPath, targetKeyPath)
	if err != nil {
		doneChan <- err
	}
	log.Infof("Sync completed successfully: s3://%s/%s -> s3://%s/%s.", sourceBucket.Name, sourceKeyPath, targetBucket.Name, targetKeyPath)
	doneChan <- nil
}

func writeS3FileToS3(sourceBucket, targetBucket *s3.Bucket, sourceKeyPath, targetKeyPath string) error {
	data, err := sourceBucket.Get(sourceKeyPath)
	if err != nil {
		return err
	}

	contType := mime.TypeByExtension(filepath.Ext(sourceKeyPath))
	Perms := s3.ACL("private")

	if err := targetBucket.Put(targetKeyPath, data, contType, Perms); err != nil {
		return err
	}

	return nil
}
