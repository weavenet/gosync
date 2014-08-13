package gosync

import "testing"

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
