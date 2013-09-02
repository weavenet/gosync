package gosync

import (
    "crypto/md5"
    "errors"
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
    Concurrent int
}

func (s *SyncPair) Sync() bool {
    if s.validPair() != true {
        fmt.Printf("Target or source not valid.\n")
        return false
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
        _, err := b.List("","","",0)
        if err == nil {
            bucket = *b
        } else if err.Error() == "Get : 301 response missing Location header" {
            continue
        } else {
            fmt.Printf("Invalid bucket.\n")
            return nil, err
        }
    }
    fmt.Printf("Found bucket in %s.\n", bucket.S3.Region.Name)
    return &bucket, nil
}

func (s *SyncPair) syncDirToS3() bool {
    sourceFiles := loadLocalFiles(s.Source)
    targetFiles, err := loadS3Files(s.Target, s.Auth)
    if err != nil {
       return false
    }

    var routines []chan string

    s3url := S3Url{Url: s.Target}
    bucket, err := lookupBucket(s3url.Bucket(), s.Auth)
    if err != nil {
       return false
    }

    count := 0
    for file, _ := range sourceFiles {
        if targetFiles[file] != sourceFiles[file] {
            count++
            filePath := strings.Join([]string{s.Source, file}, "/")
            fmt.Printf("Starting sync: %s -> s3://%s/%s.\n", filePath, bucket.Name, file)
            wait := make(chan string)
            keyPath := strings.Join([]string{s3url.Key(), file}, "/")
            go putRoutine(wait, filePath, bucket, keyPath)
            routines = append(routines, wait)
        }
        if count > s.Concurrent {
            fmt.Printf("Maxiumum concurrent threads running. Waiting.\n")
            waitForRoutines(routines)
            count = 0
            routines = routines[0:0]
        }
    }
    waitForRoutines(routines)
    return true
}

func (s *SyncPair) syncS3ToDir() bool {
    sourceFiles, err := loadS3Files(s.Source, s.Auth)
    if err != nil {
       return false
    }
    targetFiles := loadLocalFiles(s.Target)

    var routines []chan string

    s3url := S3Url{Url: s.Source}
    bucket, err := lookupBucket(s3url.Bucket(), s.Auth)
    if err != nil {
       return false
    }

    count := 0

    for file, _ := range sourceFiles {
        if targetFiles[file] != sourceFiles[file] {
            count++
            filePath := strings.Join([]string{s.Target, file}, "/")
            fmt.Printf("Starting sync: s3://%s/%s -> %s.\n", bucket.Name, file, filePath)
            if filepath.Dir(filePath) != "." {
               err := os.MkdirAll(filepath.Dir(filePath), 0755)
               if err != nil {
                  panic(err.Error())
               }
            }

            wait := make(chan string)
            go getRoutine(wait, filePath, bucket, file)
            routines = append(routines, wait)
        }
        if count > s.Concurrent {
            fmt.Printf("Maxiumum concurrent threads running. Waiting.\n")
            waitForRoutines(routines)
            count = 0
            routines = routines[0:0]
        }
    }
    waitForRoutines(routines)
    return true
}

func loadS3Files(url string, auth aws.Auth) (map[string]string, error) {
          files := map[string]string{}
          s3url := S3Url{Url: url}
          path  := s3url.Path()

          bucket, err := lookupBucket(s3url.Bucket(), auth)
          if err != nil {
             return nil, err
          }

          data, err := bucket.List(path, "", "", 0)
          if err != nil {
             panic(err.Error())
          }
          if data.IsTruncated == true {
             msg := "Results from S3 truncated and I don't yet know how to downlaod next set of results, exiting to avoid invalid results."
             fmt.Printf("%s\n", msg)
             err := errors.New(msg)
             return nil, err
          }
          for i := range data.Contents {
            md5sum := strings.Trim(data.Contents[i].ETag, "\"")
            k := relativePath(path, data.Contents[i].Key)
            files[k] = md5sum
          }
          return files, nil
}

func loadLocalFiles(path string) map[string]string {
    files := map[string]string{}
    filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
        if !info.IsDir() {
            p := relativePath(path, filePath)

            buf, err := ioutil.ReadFile(filePath)
            if err != nil {
                panic(err)
            }

            hasher := md5.New()
            hasher.Write(buf)
            md5sum := fmt.Sprintf("%x", hasher.Sum(nil))
            files[p] = md5sum
        }
        return nil
    })
    return files
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

func pathExists(path string) (bool) {
    _, err := os.Stat(path)
    if err == nil { return true }
    if os.IsNotExist(err) { return false }
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
        msg := <- r
        fmt.Printf("%s\n", msg)
    }
}

func relativePath(path string, filePath string) string {
    if path == "." {
        return strings.TrimPrefix(filePath, "/")
    } else {
        return strings.TrimPrefix(strings.TrimPrefix(filePath, path), "/")
    }
}
