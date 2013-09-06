package gosync

import "testing"

type validS3UrlTestCase struct {
  url    string
  result bool
}

var validS3UrlTests = []validS3UrlTestCase{
  {"s3://bucket/test.tar.gz", true},
  {"s3://bucket/test/123", true},
  {"s3://bucket/123", true},
  {"s3://bucket", true},
  {"bucket", false},
}

func TestValidS3UrlTests(t *testing.T) {
  for _, c := range validS3UrlTests {
    if validS3Url(c.url) != c.result {
      t.Error("Validation failed.")
    }
  }
}

type relativePathTestCase struct {
  path     string
  filePath string
  result   string
}

var relativePathTests = []relativePathTestCase{
  {"/home/me", "/home/me/my/file", "my/file"},
  {".", "/my/file", "my/file"},
}

func TestRelativePaths(t *testing.T) {
  for _, c := range relativePathTests {
    if relativePath(c.path, c.filePath) != c.result {
      t.Error("Relative path returned incorrectly.")
    }
  }
}
