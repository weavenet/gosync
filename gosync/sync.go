package gosync

import (
    "crypto/md5"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "launchpad.net/goamz/aws"
    "launchpad.net/goamz/s3"
)

type SyncPair struct {
    Source string
    Target string
    Auth aws.Auth
}

func (s *SyncPair) Sync() bool {
    if s.validPair() {
        if validS3Url(s.Source) {
           if s.syncS3ToDir() == true {
               return true
           }
        } else {
           if s.syncDirToS3() == true {
               return true
           }
        }
    }
    fmt.Printf("Path not valid.")
    return false
}

func (s *SyncPair) syncDirToS3() bool {
    sourceFiles := loadLocalFiles(s.Source)
    targetFiles := loadS3Files(s.Target, s.Auth)

    region := aws.USEast
    s3 := s3.New(s.Auth, region)
    s3url := S3Url{Url: s.Target}

    for file, _ := range sourceFiles {
        if targetFiles[file] != sourceFiles[file] {
            filePath := strings.Join([]string{s.Source, file}, "/")
            bucket := s3.Bucket(s3url.Bucket())
            fmt.Printf("Syncing %s to %s in bucket %s.\n", filePath, file, bucket.Name)
            Put(bucket, file, filePath)
        }
    }
    return true
}

func (s *SyncPair) syncS3ToDir() bool {
    sourceFiles := loadS3Files(s.Source, s.Auth)
    targetFiles := loadLocalFiles(s.Target)

    region := aws.USEast
    s3 := s3.New(s.Auth, region)
    s3url := S3Url{Url: s.Source}

    for file, _ := range sourceFiles {
        if targetFiles[file] != sourceFiles[file] {
            filePath := strings.Join([]string{s.Target, file}, "/")
            bucket := s3.Bucket(s3url.Bucket())
            fmt.Printf("Syncing %s from bucket %s to %s.\n", file, bucket.Name, filePath)
            Get(filePath, bucket, file)
        }
    }
    return true
}

func loadS3Files(url string, auth aws.Auth) map[string]string {
    files := map[string]string{}
          s3url := S3Url{Url: url}
          key := s3url.Key()
          region := aws.USEast
          s := s3.New(auth, region)
          bucket := s.Bucket(s3url.Bucket())
          defer func() {
              if r := recover(); r != nil {
                  fmt.Printf("%v", r)
              }
          }()
          data, err := bucket.List(key, "", "", 0)
          if err != nil {
             panic(err.Error())
          }
          for i := range data.Contents {
            md5sum := strings.Trim(data.Contents[i].ETag, "\"")
            k := strings.TrimPrefix(data.Contents[i].Key, url)
            fmt.Printf("Read sum from S3 file %s with md5sum %s\n", k, md5sum)
            files[k] = md5sum
          }
          return files
}

func loadLocalFiles(path string) map[string]string {
    files := map[string]string{}
    filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
        if !info.IsDir() {
            relativePath := strings.TrimPrefix(filePath, path)
            fmt.Printf("For some reason not loading all files.  Saving: %s\n", relativePath)

            buf, err := ioutil.ReadFile(filePath)
            if err != nil {
                panic(err)
            }

            hasher := md5.New()
            hasher.Write(buf)
            md5sum := fmt.Sprintf("%x", hasher.Sum(nil))
            fmt.Printf("Read sum from local file %s with md5sum %s\n", relativePath, md5sum)
            files[relativePath] = md5sum
        }
        return nil
    })
    return files
}

func (s *SyncPair) validPair() bool {
     if pathExists(s.Source) == false && pathExists(s.Target) == false {
         return false
     }
     if validS3Url(s.Source) == false && validS3Url(s.Target) == false {
         return false
     }
     return true
}

func validS3Url(path string) bool {
    return strings.HasPrefix(path, "s3://")
}

func pathExists(path string) (bool) {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
    return false
}
