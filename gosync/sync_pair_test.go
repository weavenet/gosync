package gosync

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/goamz/aws"
)

func TestValidSyncPair(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "dir")
	if err != nil {
		t.Fatalf("Error creating temp dir")
	}

	tcDir1 := tempDir + "/dir1"
	tcDir2 := tempDir + "/dir2"

	for _, d := range []string{tcDir1, tcDir2} {
		if err := os.Mkdir(d, 0755); err != nil {
			t.Fatalf("Error creating temp file")
		}
	}

	var syncPairTCs = []struct {
		source string
		target string
		valid  bool
	}{
		{"s3://b1", "s3://b2", true},
		{tcDir1, "s3://b2", true},
		{"s3://b1", tcDir2, true},
		{tcDir1, tcDir2, false},
		{"s3://b1", tempDir + "/bad_dir", false},
	}

	for _, tc := range syncPairTCs {
		auth := aws.Auth{}
		sp := NewSyncPair(auth, tc.source, tc.target, "regionX")
		if sp.validPair() != tc.valid {
			t.Fatalf("Error testing sync pair validity for %s -> %s", tc.source, tc.target)
		}
	}
}
