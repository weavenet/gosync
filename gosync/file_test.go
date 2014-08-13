package gosync

import (
	"io/ioutil"
	"testing"
)

type relativePathTestCase struct {
	path     string
	filePath string
	result   string
}

func TestLoadLocalFiles(t *testing.T) {
	dir, err := ioutil.TempDir("", "dir")
	if err != nil {
		t.Fatalf("Error creating temp dir")
	}

	if err := ioutil.WriteFile(dir+"/file", []byte("test1234"), 0400); err != nil {
		t.Fatalf("Error creating temp file")
	}

	data, err := loadLocalFiles(dir)
	if err != nil {
		t.Fatalf("Received error loading local files.")
	}

	if data["file"] != "16d7a4fca7442dda3ad93c9a726597e4" {
		t.Fatalf("Data not correctly load from local files.")
	}
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
