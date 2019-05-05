package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
)

// UnTar extracts all files from the given stream (tar file) into the given path
func UnTar(tarstream io.Reader, dst Path) error {
	tarReader := tar.NewReader(tarstream)

	type symlink struct {
		old string
		new string
	}

	symlinks := make([]symlink, 0)
	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("ExtractTar: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(dst.Add(Path(header.Name)).String(), 0755); err != nil {
				return fmt.Errorf("ExtractTar: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(dst.Add(Path(header.Name)).String(), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("ExtractTar: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				_ = outFile.Close()
				return fmt.Errorf("ExtractTar: Copy() failed: %s", err.Error())
			}
			_ = outFile.Close()
		case tar.TypeSymlink:
			src := dst.Add(Path(header.Name))
			symlinks = append(symlinks, symlink{old: src.String(), new: header.Linkname})

		default:
			return fmt.Errorf("ExtractTar: unknown type: %d (%s) in %s", header.Typeflag, string(header.Typeflag), header.Name)
		}
	}

	// postpone our symlink creation
	for _, symlink := range symlinks {

		// for the jdk, the meaning of .. seems to be broken. The information e.g. about the extracted jdk8u212-b03
		// directory is lost, perhaps also a bug in the go tar reader?

		logger.Error(Fields{"action": "symlink", "src": symlink.old, "dst": symlink.new})

	}
	return nil
}
