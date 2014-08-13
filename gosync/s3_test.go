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
