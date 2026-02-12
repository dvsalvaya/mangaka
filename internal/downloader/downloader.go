package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type Downloader struct {
	client *http.Client
}

func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{},
	}
}

func (d *Downloader) DownloadAndRead(chapterID string, urls []string) error {
	tmpDir := filepath.Join(os.TempDir(), "mangaka", chapterID)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up raw files after CBZ creation

	// 1. Download Images
	var wg sync.WaitGroup
	errChan := make(chan error, len(urls))
	files := make([]string, len(urls))

	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			ext := filepath.Ext(url)
			if ext == "" {
				ext = ".jpg"
			}
			filename := fmt.Sprintf("%03d%s", i+1, ext) // 001.jpg
			path := filepath.Join(tmpDir, filename)

			if err := d.downloadFile(url, path); err != nil {
				errChan <- err
				return
			}
			files[i] = path // Store for zipping order
		}(i, url)
	}

	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		return <-errChan
	}

	// 2. Create CBZ
	cbzPath := filepath.Join(os.TempDir(), fmt.Sprintf("mangaka_chapter_%s.cbz", chapterID))
	if err := d.createCBZ(cbzPath, files); err != nil {
		return err
	}

	// 3. Open Zathura
	fmt.Printf("Opening %s with Zathura...\n", cbzPath)
	cmd := exec.Command("zathura", cbzPath)
	return cmd.Start()
}

func (d *Downloader) downloadFile(url, path string) error {
	resp, err := d.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (d *Downloader) createCBZ(cbzPath string, files []string) error {
	cbzFile, err := os.Create(cbzPath)
	if err != nil {
		return err
	}
	defer cbzFile.Close()

	zipWriter := zip.NewWriter(cbzFile)
	defer zipWriter.Close()

	for _, file := range files {
		if file == "" {
			continue
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// Add to zip
		w, err := zipWriter.Create(filepath.Base(file))
		if err != nil {
			f.Close()
			return err
		}

		_, err = io.Copy(w, f)
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
