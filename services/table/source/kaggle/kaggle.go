package kaggle

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func download(dir string, url string) error {
	resp, err := http.Get(url) //nolint
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	out, err := os.Create(filepath.Join(dir, "download.zip"))
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, resp.Body)
	return err
}

// https://snyk.io/research/zip-slip-vulnerability
// https://github.com/securego/gosec/issues/324#issuecomment-935927967
func sanitizeArchivePath(dest string, filename string) (v string, err error) {
	v = filepath.Join(dest, filename)
	if strings.HasPrefix(v, filepath.Clean(dest)) {
		return v, nil
	}

	return "", fmt.Errorf("%s: %s", "content filepath is tainted", filename)
}

func unzip(src string, dest string) error {
	zipReader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		filePath, err := sanitizeArchivePath(dest, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			return err
		}

		outFile, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		zipFile, err := file.Open()
		if err != nil {
			return err
		}
		defer zipFile.Close()
		for {
			_, err := io.CopyN(outFile, zipFile, 2<<20)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
		}
	}
	return nil
}

// PrepareKaggleDataset download zip file using kaggle API and unzip to tablepilot_kaggle_cache folder
func PrepareKaggleDataset(ctx context.Context, dataset string, dir string) (string, error) {
	path := strings.ReplaceAll(dataset, "/", "--")
	cachePath := filepath.Join(dir, "tablepilot_kaggle_cache", path)
	zipPath := filepath.Join(dir, "tablepilot_kaggle_cache/tmp", path)
	_, err := os.Stat(cachePath)
	// cache exists
	if err == nil {
		return "", nil
	}

	// download kaggle dataset zip
	url := fmt.Sprintf("https://www.kaggle.com/api/v1/datasets/download/%s", dataset)
	err = download(zipPath, url)
	defer func() { _ = os.RemoveAll(zipPath) }()
	if err != nil {
		return "", err
	}
	return cachePath, unzip(filepath.Join(zipPath, "download.zip"), cachePath)
}
