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
	trimmed_string := strings.TrimPrefix(r.Url, "s3://")
	return strings.Split(trimmed_string, "/")
}

func loadS3Files(bucket *s3.Bucket, path string, files map[string]string, marker string) (map[string]string, error) {
	log.Debugf("Loading files from 's3://%s/%s'.", bucket.Name, path)
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

	log.Debugf("Loaded '%d' files from 's3://%s/%s' succesfully.", len(files), bucket.Name, path)
	return files, nil
}

func lookupBucket(bucketName string, auth aws.Auth, region string) (*s3.Bucket, error) {
	log.Infof("Looking up region for bucket '%s'.", bucketName)

	if(region != "") {
		log.Debugf("Looking for bucket '%s' in '%s'.", bucketName, region)
		s3 := s3.New(auth, aws.Regions[region])
		bucket := s3.Bucket(bucketName)

		// If list return, bucket is valid in this region.
		_, err := bucket.List("", "", "", 0)
		if err == nil {
			log.Infof("Found bucket '%s' in '%s'.", bucketName, region)
			return bucket, nil
		}
	}

	// Looking in each region for bucket
	// To do, make this less crusty and ghetto

	for lregion, _ := range aws.Regions {
		// Current does not support gov lregion or china
		if lregion == "us-gov-west-1" || lregion == "cn-north-1" {
			log.Debugf("Skipping %s", lregion)
			continue
		}

		log.Debugf("Looking for bucket '%s' in '%s'.", bucketName, lregion)
		s3 := s3.New(auth, aws.Regions[lregion])
		bucket := s3.Bucket(bucketName)

		// If list return, bucket is valid in this lregion.
		_, err := bucket.List("", "", "", 0)
		if err == nil {
			log.Infof("Found bucket '%s' in '%s'.", bucketName, lregion)
			return bucket, nil
		} else if strings.Contains(err.Error(), "301 response missing Location header") {
			log.Debugf("Bucket '%s' not found in '%s'.", bucketName, lregion)
			continue
		} else {
			return nil, err
		}
	}

	return nil, fmt.Errorf("Bucket not found.")
}
