package gosync

import (
  "io/ioutil"
  "launchpad.net/goamz/s3"
  "os"
  "strings"
)

type S3Url struct {
  Url string
}

func (r *S3Url) Bucket() string {
  return r.keys()[0]
}

func (r *S3Url) Key() string {
  return strings.Join(r.keys()[1:len(r.keys())], "/")
}

func (r *S3Url) Path() string {
  return r.Key()
}

func (r *S3Url) Valid() bool {
  return strings.HasPrefix(r.Url, "s3://")
}

func (r *S3Url) keys() []string {
  trimmed_string := strings.TrimLeft(r.Url, "s3://")
  return strings.Split(trimmed_string, "/")
}

func Get(file string, bucket *s3.Bucket, path string) {
  data, err := bucket.Get(path)
  if err != nil {
    panic(err.Error())
  }
  perms := os.FileMode(0644)

  err = ioutil.WriteFile(file, data, perms)
  if err != nil {
    panic(err.Error())
  }
}

func Put(bucket *s3.Bucket, path string, file string) {
  contType := "binary/octet-stream"
  Perms := s3.ACL("private")

  data, err := ioutil.ReadFile(file)
  if err != nil {
    panic(err.Error())
  }

  err = bucket.Put(path, data, contType, Perms)
  if err != nil {
    panic(err.Error())
  }
}
