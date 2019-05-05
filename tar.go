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
		default:
			return fmt.Errorf("ExtractTar: uknown type: %d in %s", header.Typeflag, header.Name)
		}
	}
	return nil
}
