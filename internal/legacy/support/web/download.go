package web

import (
	"fmt"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"net/http"
	"os"
	"strings"
)

func newDownloadErr(uri string, msg string) error {
	return fmt.Errorf("failed to download from '%s': %s", uri, msg)
}

// ProgressDownloadFile downloads specified file from URL and displaying download progress in terminal
func ProgressDownloadFile(client *http.Client, uri, destination string) error {
	out, err := os.Create(destination + ".tmp")
	if err != nil {
		return err
	}

	defer out.Close()
	resp, err := client.Get(uri)
	if err != nil {
		return newDownloadErr(uri, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return newDownloadErr(uri, resp.Status)
	}

	total := int(resp.ContentLength)
	bar := pb.StartNew(total)
	bar.SetUnits(pb.U_BYTES)
	reader := bar.NewProxyReader(resp.Body)
	defer func() {
		bar.NotPrint = true
		bar.Finish()

		// Put cursor to the beginning and overwrite progress bar line with empty spaces
		// to remove the bar
		fmt.Fprint(os.Stdout, "\r\b", strings.Repeat(" ", bar.GetWidth()), "\r")
	}()

	_, err = io.Copy(out, reader)
	if err != nil {
		return fmt.Errorf("failed to save downloaded file: %s", err)
	}

	return os.Rename(out.Name(), destination)
}
