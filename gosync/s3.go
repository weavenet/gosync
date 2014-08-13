package gosync

import (
	"fmt"
	"strings"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type s3Url struct {
	Url string
}

func newS3Url(url string) s3Url {
	return s3Url{Url: url}
}

func (r *s3Url) Bucket() string {
	return r.keys()[0]
}

func (r *s3Url) Key() string {
	return strings.Join(r.keys()[1:len(r.keys())], "/")
}

func (r *s3Url) Path() string {
	return r.Key()
}

func (r *s3Url) Valid() bool {
	return strings.HasPrefix(r.Url, "s3://")
}

func (r *s3Url) keys() []string {
	trimmed_string := strings.TrimLeft(r.Url, "s3://")
	return strings.Split(trimmed_string, "/")
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
