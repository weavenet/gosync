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

type SyncPair struct {
	Source     string
	Target     string
	Auth       aws.Auth
	Concurrent int
}

func (s *SyncPair) Sync() error {
	if s.validPair() != true {
		return errors.New("Invalid sync pair.")
	}

	if validS3Url(s.Source) {
		return s.syncS3ToDir()
	} else {
		return s.syncDirToS3()
	}
}

func lookupBucket(bucketName string, auth aws.Auth) (*s3.Bucket, error) {
	var bucket s3.Bucket

	// Looking in each region for bucket
	// To do, make this less crusty and ghetto
	for r, _ := range aws.Regions {
		s3 := s3.New(auth, aws.Regions[r])
		b := s3.Bucket(bucketName)

		// If list return, bucket is valid in this region.
		_, err := b.List("", "", "", 0)
		if err == nil {
			bucket = *b
		} else if err.Error() == "Get : 301 response missing Location header" {
			continue
		} else {
			return nil, fmt.Errorf("Invalid bucket.\n")
		}
	}
	log.Infof("Found bucket in '%s'.", bucket.S3.Region.Name)
	return &bucket, nil
}

func (s *SyncPair) syncDirToS3() error {
	log.Infof("Syncing to S3.")

	sourceFiles, err := loadLocalFiles(s.Source)
	if err != nil {
		return err
	}

	s3url := S3Url{Url: s.Target}
	path := s3url.Path()

	bucket, err := lookupBucket(s3url.Bucket(), s.Auth)
	if err != nil {
		return err
	}

	// Load files and do not specify marker to start
	targetFiles := make(map[string]string)
	targetFiles, err = loadS3Files(bucket, path, targetFiles, "")
	if err != nil {
		return err
	}

	var routines []chan string

	count := 0
	for file, _ := range sourceFiles {
		if targetFiles[file] != sourceFiles[file] {
			count++
			filePath := strings.Join([]string{s.Source, file}, "/")
			log.Infof("Starting sync: %s -> s3://%s/%s.", filePath, bucket.Name, file)
			wait := make(chan string)
			keyPath := strings.Join([]string{s3url.Key(), file}, "/")
			go putRoutine(wait, filePath, bucket, keyPath)
			routines = append(routines, wait)
		}
		if count > s.Concurrent {
			log.Infof("Maxiumum concurrent threads running. Waiting.")
			waitForRoutines(routines)
			count = 0
			routines = routines[0:0]
		}
	}
	waitForRoutines(routines)
	return nil
}

func (s *SyncPair) syncS3ToDir() error {
	log.Infof("Syncing from S3.")

	s3url := S3Url{Url: s.Source}
	bucket, err := lookupBucket(s3url.Bucket(), s.Auth)
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

	var routines []chan string

	count := 0

	for file, _ := range sourceFiles {
		if targetFiles[file] != sourceFiles[file] {
			count++
			filePath := strings.Join([]string{s.Target, file}, "/")
			log.Infof("Starting sync: s3://%s/%s -> %s.", bucket.Name, file, filePath)
			if filepath.Dir(filePath) != "." {
				err := os.MkdirAll(filepath.Dir(filePath), 0755)
				if err != nil {
					return err
				}
			}

			wait := make(chan string)
			go getRoutine(wait, filePath, bucket, file)
			routines = append(routines, wait)
		}
		if count > s.Concurrent {
			log.Infof("Maxiumum concurrent threads running. Waiting...")
			waitForRoutines(routines)
			count = 0
			routines = routines[0:0]
		}
	}
	waitForRoutines(routines)
	return nil
}

func loadS3Files(bucket *s3.Bucket, path string, files map[string]string, marker string) (map[string]string, error) {
	data, err := bucket.List(path, "", marker, 0)
	if err != nil {
		return files, err
	}

	for i := range data.Contents {
		md5sum := strings.Trim(data.Contents[i].ETag, "\"")
		k := relativePath(path, data.Contents[i].Key)
		files[k] = md5sum
	}

	// Continue to call loadS3files if next marker set
	if data.IsTruncated {
		lastKey := data.Contents[(len(data.Contents) - 1)].Key
		log.Infof("Results truncated, loading additional files via previous last key '%s'.", lastKey)
		loadS3Files(bucket, path, files, lastKey)
	} else {
		log.Infof("All keys loaded.")
	}
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

	return files, err
}

func (s *SyncPair) validPair() bool {
	if validTarget(s.Source) == true && validTarget(s.Target) == true {
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

func putRoutine(quit chan string, filePath string, bucket *s3.Bucket, file string) {
	Put(bucket, file, filePath)
	quit <- fmt.Sprintf("Completed sync: %s -> s3://%s/%s.", filePath, bucket.Name, file)
}

func getRoutine(quit chan string, filePath string, bucket *s3.Bucket, file string) {
	Get(filePath, bucket, file)
	quit <- fmt.Sprintf("Completed sync: s3://%s/%s -> %s.", bucket.Name, file, filePath)
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
