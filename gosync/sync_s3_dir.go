package gosync

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/goamz/s3"
)

func (s *SyncPair) syncS3ToDir() error {
	log.Infof("Syncing from S3.")

	s3url := newS3Url(s.Source)
	bucket, err := lookupBucket(s3url.Bucket(), s.Auth, s.Region)
	if err != nil {
		return err
	}

	sourceFiles := make(map[string]string)
	sourceFiles, err = loadS3Files(bucket, s3url.Path(), sourceFiles, "")
	if err != nil {
		return err
	}

	targetFiles, err := loadLocalFiles(s.Target)
	if err != nil {
		return err
	}
	return s.concurrentSyncS3ToDir(s3url, bucket, targetFiles, sourceFiles)
}

func (s *SyncPair) concurrentSyncS3ToDir(s3url s3Url, bucket *s3.Bucket, targetFiles, sourceFiles map[string]string) error {
	doneChan := newDoneChan(s.Concurrent)
	pool := newPool(s.Concurrent)
	var wg sync.WaitGroup

	for file, _ := range sourceFiles {
		if targetFiles[file] != sourceFiles[file] {
			filePath := strings.Join([]string{s.Target, file}, "/")
			if filepath.Dir(filePath) != "." {
				err := os.MkdirAll(filepath.Dir(filePath), 0755)
				if err != nil {
					return err
				}
			}

			// Get transfer reservation from pool
			log.Tracef("Requesting reservation for '%s'.", filePath)
			<-pool
			log.Tracef("Retrieved reservation for '%s'.", filePath)

			log.Infof("Starting sync: s3://%s/%s -> %s.", bucket.Name, file, filePath)
			wg.Add(1)
			go func(doneChan chan error, filePath string, bucket *s3.Bucket, file string) {
				defer wg.Done()
				writeS3FileToPathRoutine(doneChan, filePath, bucket, file)
				pool <- 1
			}(doneChan, filePath, bucket, file)
		}
	}

	wg.Wait()
	return nil
}

func writeS3FileToPathRoutine(doneChan chan error, filePath string, bucket *s3.Bucket, file string) {
	err := writeS3FileToPath(filePath, bucket, file)
	if err != nil {
		doneChan <- err
	}
	log.Infof("Sync completed successfully: s3://%s/%s -> %s.", bucket.Name, file, filePath)
	doneChan <- nil
}

func writeS3FileToPath(file string, bucket *s3.Bucket, path string) error {
	data, err := bucket.Get(path)
	if err != nil {
		return err
	}
	perms := os.FileMode(0644)

	err = ioutil.WriteFile(file, data, perms)
	if err != nil {
		return err
	}

	return nil
}
