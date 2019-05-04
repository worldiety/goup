package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
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
func mkdirs(fname Path) error{
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
