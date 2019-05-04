package main

// A Downloader downloads stuff like a zip folder and unpacks it into a the context of GoUp
type Downloader interface {
	// Performs the Download, or fails
	Download(goup *Goup) error
}

type GoDownloader struct {
}

func (GoDownloader) Download(gp *Goup) error {
	return nil
}
