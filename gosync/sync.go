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

func (s *SyncPair) syncDirToS3() bool {
    sourceFiles := loadLocalFiles(s.Source)
    targetFiles := loadS3Files(s.Target, s.Auth)

    region := aws.USEast
    s3 := s3.New(s.Auth, region)
    s3url := S3Url{Url: s.Target}

    var routines []chan string

    count := 0
    for file, _ := range sourceFiles {
        if targetFiles[file] != sourceFiles[file] {
            count++
            filePath := strings.Join([]string{s.Source, file}, "/")
            bucket := s3.Bucket(s3url.Bucket())
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
    sourceFiles := loadS3Files(s.Source, s.Auth)
    targetFiles := loadLocalFiles(s.Target)

    region := aws.USEast
    s3 := s3.New(s.Auth, region)
    s3url := S3Url{Url: s.Source}

    var routines []chan string

    count := 0

    for file, _ := range sourceFiles {
        if targetFiles[file] != sourceFiles[file] {
            count++
            filePath := strings.Join([]string{s.Target, file}, "/")
            bucket := s3.Bucket(s3url.Bucket())
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

func loadS3Files(url string, auth aws.Auth) map[string]string {
          files := map[string]string{}
          s3url := S3Url{Url: url}
          path  := s3url.Path()

          region := aws.USEast
          s := s3.New(auth, region)
          bucket := s.Bucket(s3url.Bucket())

          data, err := bucket.List(path, "", "", 0)
          if err != nil {
             panic(err.Error())
          }
          for i := range data.Contents {
            md5sum := strings.Trim(data.Contents[i].ETag, "\"")
            k := strings.TrimPrefix(data.Contents[i].Key, url)
            files[k] = md5sum
          }
          return files
}

func loadLocalFiles(path string) map[string]string {
    files := map[string]string{}
    filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
        if !info.IsDir() {
            var relativePath string

            if path == "." {
                relativePath = filePath
            } else {
                relativePath = strings.TrimPrefix(strings.TrimPrefix(filePath, path), "/")
            }

            buf, err := ioutil.ReadFile(filePath)
            if err != nil {
                panic(err)
            }

            hasher := md5.New()
            hasher.Write(buf)
            md5sum := fmt.Sprintf("%x", hasher.Sum(nil))
            files[relativePath] = md5sum
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
    if pathExists(target) {
        return true
    }
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
