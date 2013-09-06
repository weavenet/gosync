package gosync

import "testing"

type S3UrlTestCase struct {
  url    string
  bucket string
  key    string
  valid  bool
}

var S3UrlTests = []S3UrlTestCase{
  {"s3://bucket/test.tar.gz", "bucket", "test.tar.gz", true},
  {"s3://bucket-123/dir/folder/key", "bucket-123", "dir/folder/key", true},
  {"s3://bucket-123/files*", "bucket-123", "files*", true},
  {"bucket-123/dir/folder/key", "bucket-123", "dir/folder/key", false},
  {"bucket-123", "bucket-123", "", false},
}

func TestS3Url(t *testing.T) {
  for _, c := range S3UrlTests {
    url := S3Url{Url: c.url}

    if url.Key() != c.key {
      t.Error("Key not returned correctly.")
    }

    if url.Bucket() != c.bucket {
      t.Error("Bucket not returned correctly.")
    }

    if url.Valid() != c.valid {
      t.Error("Validation did not return correctly.")
    }
  }
}
