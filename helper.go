package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//CWD returns the current working dir or panics if it is unknown
func CWD() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return dir
}

// mkdirs ensures the path existence
func mkdirs(fname Path) error {
	return os.MkdirAll(fname.String(), os.ModePerm)
}

// must terminates the process if err is not nil
func must(err error) {
	if err != nil {
		logger.Error(Fields{"err": err})
		os.Exit(-1)
	}
}

// getModuleName parses the given file denoted by fname and returns the declared moduleName name
func getModuleName(fname Path) (string, error) {
	data, err := ioutil.ReadFile(fname.String())
	if err != nil {
		return "", err
	}
	str := string(data)
	lines := strings.Split(str, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module") {
			modName := strings.TrimSpace(line[len("module"):])
			modName = strings.Replace(modName, "\"", "", -1)
			return modName, nil
		}
	}
	return "", fmt.Errorf("go.mod does not contain a module definition")
}

// CopyDir copies a whole directory recursively
func CopyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// CopyFile copies a single file from src to dst
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func ListFiles(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}

		}

		files = append(files, path)
		return nil
	})
	return files, err
}

// IsEmpty returns true if given string only consists of whitespace chars or is empty
func IsEmpty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// Sha256 calculates a Sha256 hash from the given stromg
func Sha256(str string) string {
	t := sha256.Sum256([]byte(str))
	return hex.EncodeToString(t[:])
}

// Downloads a large file without memory buffering
func DownloadFile(url string, dstFile string) error {
	logger.Debug(Fields{"action": "downloading", "url": url, "dst": dstFile})

	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cannot download, bad status: %s", resp.Status)
	}

	lastPrinted := time.Now()
	pgReader := &progressReader{resp.ContentLength, 0, func(read int64, max int64) {
		p := int(float64(read) / float64(max) * 100)
		if time.Now().Sub(lastPrinted).Seconds() > 15 {
			lastPrinted = time.Now()
			logger.Info(Fields{"action": "progress", "status": strconv.Itoa(p) + "%"})
		}
	}, resp.Body}

	// Writer the body to file
	_, err = io.Copy(out, pgReader)
	if err != nil {
		return err
	}

	logger.Debug(Fields{"action": "completed", "url": url, "dst": dstFile})
	return nil
}

func downloadAndUnpack(url string, targetFolder Path) error {
	tmpFile := targetFolder.Parent().Child(Sha256(url) + ".tmp")
	defer os.Remove(tmpFile.String())

	err := DownloadFile(url, tmpFile.String())
	if err != nil {
		return err
	}
	srcFile, err := os.OpenFile(tmpFile.String(), os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	lname := strings.ToLower(url)
	if strings.HasSuffix(lname, ".tar.gz") {
		uncompressedStream, err := gzip.NewReader(srcFile)
		if err != nil {
			return fmt.Errorf("gz stream failed: %v", err)
		}
		return UnTar(uncompressedStream, targetFolder)
	}

	if strings.HasSuffix(lname, ".zip") {
		return Unzip(tmpFile.String(), targetFolder.String())
	}

	return fmt.Errorf("unsupported file format: %s", filepath.Ext(lname))
}

type progressReader struct {
	max      int64
	current  int64
	callback func(read int64, max int64)
	delegate io.Reader
}

func (r *progressReader) Read(p []byte) (n int, err error) {
	n, err = r.delegate.Read(p)
	r.current += int64(n)
	if r.callback != nil {
		r.callback(r.current, r.max)
	}
	return n, err
}

