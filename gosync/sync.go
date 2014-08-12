package gosync

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
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

	//if validS3Url(s.Source) && validS3Url(s.Target) {
	//	return s.syncS3ToS3()
	//}

	if validS3Url(s.Source) {
		return s.syncS3ToDir()
	}

	return s.syncDirToS3()
}

func lookupBucket(bucketName string, auth aws.Auth) (*s3.Bucket, error) {
	log.Infof("Looking up region for bucket '%s'.", bucketName)

	var bucket *s3.Bucket = nil

	// Looking in each region for bucket
	// To do, make this less crusty and ghetto
	for region, _ := range aws.Regions {
		log.Debugf("Looking for bucket '%s' in '%s'.", bucketName, region)
		s3 := s3.New(auth, aws.Regions[region])
		b := s3.Bucket(bucketName)

		// If list return, bucket is valid in this region.
		_, err := b.List("", "", "", 0)
		if err == nil {
			log.Infof("Found bucket '%s' in '%s'.", bucketName, region)
			bucket = b
			break
		} else if err.Error() == "Get : 301 response missing Location header" {
			log.Debugf("Bucket '%s' not found in '%s'.", bucketName, region)
			continue
		} else {
			return nil, err
		}
	}

	if bucket != nil {
		return bucket, nil
	}

	return nil, fmt.Errorf("Bucket not found.")
}

func loadS3Files(bucket *s3.Bucket, path string, files map[string]string, marker string) (map[string]string, error) {
	data, err := bucket.List(path, "", marker, 0)
	if err != nil {
		return files, err
	}

	for _, key := range data.Contents {
		md5sum := strings.Trim(key.ETag, "\"")
		files[key.Key] = md5sum
	}

	// Continue to call loadS3files and add
	// Files to map if next marker set
	if data.IsTruncated {
		lastKey := data.Contents[(len(data.Contents) - 1)].Key
		log.Infof("Results truncated, loading additional files via previous last key '%s'.", lastKey)
		loadS3Files(bucket, path, files, lastKey)
	}

	log.Debugf("Loaded '%d' files from S3.", len(files))
	log.Infof("Loading files from S3 complete.")
	return files, nil
}

func loadLocalFiles(path string) (map[string]string, error) {
	files := map[string]string{}

	loadMd5Sums := func(filePath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			p := relativePath(path, filePath)

			buf, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			hasher := md5.New()
			hasher.Write(buf)
			md5sum := fmt.Sprintf("%x", hasher.Sum(nil))
			files[p] = md5sum
		}
		return nil
	}

	err := filepath.Walk(path, loadMd5Sums)
	if err != nil {
		return files, err
	}

	log.Debugf("Loaded '%d' files from '%s'.", len(files), path)
	log.Infof("Loading local files complete.")

	return files, nil
}

func (s *Sync) validPair() bool {
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

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
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
