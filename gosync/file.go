package gosync

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/cihub/seelog"
)

func loadLocalFiles(path string) (map[string]string, error) {
	files := map[string]string{}

	loadMd5Sums := func(filePath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			p := relativePath(path, filePath)

			buf, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			hasher := md5.New()
			hasher.Write(buf)
			md5sum := fmt.Sprintf("%x", hasher.Sum(nil))
			files[p] = md5sum
		}
		return nil
	}

	err := filepath.Walk(path, loadMd5Sums)
	if err != nil {
		return files, err
	}

	log.Debugf("Loaded '%d' files from '%s'.", len(files), path)
	log.Infof("Loading local files complete.")

	return files, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
