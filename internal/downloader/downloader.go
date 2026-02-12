package downloader

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dvsalvaya/mangaka/pkg/opener"
)

const DownloadDir = "downloads"

type Downloader struct {
	client  *http.Client
	BaseDir string
}

func NewDownloader() *Downloader {
	return &Downloader{
		client:  &http.Client{},
		BaseDir: DownloadDir,
	}
}

func (d *Downloader) DownloadAndRead(chapterID string, urls []string) error {
	// 1. Setup Temp Dirs
	tmpBase := filepath.Join(os.TempDir(), "mangaka_read")
	if err := os.MkdirAll(tmpBase, 0755); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Create a unique folder for this read session
	sessionID := fmt.Sprintf("%s_%d", chapterID, time.Now().Unix())
	imgDir := filepath.Join(tmpBase, sessionID, "images")
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return fmt.Errorf("failed to create image dir: %w", err)
	}

	// Ensure cleanup - though if we open async, we might delete before reading?
	// opener.Open returns immediately.
	// If we delete immediately, the viewer might fail.
	// Improved strategy: Don't delete immediately. Rely on OS temp cleanup or
	// delete heavily old folders on startup.
	// For CLI simplicity, we'll leave them for now or delete on exit if we could block.
	// But `opener.Start` is non-blocking usually.
	// Let's NOT defer remove for now to ensure viewer can open it.
	// Ideally we'd have a "Session Manager" dealing with this.
	// A simple approach: Delete previous sessions on startup, not now.

	// 2. Progressive Download & Open
	if len(urls) > 0 {
		// A. Download First Page Priority
		firstURL := urls[0]
		ext := filepath.Ext(firstURL)
		if ext == "" {
			ext = ".jpg"
		}
		firstFilename := fmt.Sprintf("%03d%s", 1, ext)
		firstPath := filepath.Join(imgDir, firstFilename)

		fmt.Printf("Downloading first page...\n")
		if err := d.downloadFile(firstURL, firstPath); err != nil {
			return fmt.Errorf("failed to download first page: %w", err)
		}

		// B. Open Immediately
		// Note: On Windows 'start' allows opening the file in default viewer.
		// Opening the image directly.
		fmt.Printf("Opening %s...\n", firstPath)
		if err := opener.Open(firstPath); err != nil {
			fmt.Printf("Warning: failed to open viewer: %v\n", err)
		}

		// C. Download the rest
		if len(urls) > 1 {
			fmt.Printf("Downloading remaining %d pages in background...\n", len(urls)-1)
			remainingURLs := urls[1:]
			// We pass an offset to file naming
			// Actually downloadImages takes a listing.
			// Let's manually handle the loop to respect the naming or refactor downloadImages.
			// To keep it simple, let's just loop here or create a helper that takes an index offset.
			// Reusing downloadImages is tricky because of the index.
			// Let's just inline the parallel logic here or make a helper.

			// We block here until done to ensure process doesn't exit if CLI logic changes,
			// and to show progress if we wanted.
			d.downloadBatch(remainingURLs, imgDir, 1) // offset 1
		}
	}

	// We do NOT create a CBZ for "Online Read" anymore because we are opening the raw image
	// to allow progressive reading.
	// The temp folder structure is: temp/session/images/001.jpg...

	return nil
}

// Helper for batch download with offset
func (d *Downloader) downloadBatch(urls []string, dir string, offset int) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(urls))

	for i, url := range urls {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			ext := filepath.Ext(url)
			if ext == "" {
				ext = ".jpg"
			}
			// index is i + offset + 1 (because 0-based i)
			filename := fmt.Sprintf("%03d%s", i+offset+1, ext)
			path := filepath.Join(dir, filename)

			if err := d.downloadFile(url, path); err != nil {
				errChan <- err
				return
			}
		}(i, url)
	}
	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		return <-errChan
	}
	return nil
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

func (d *Downloader) DownloadChapter(chapterID string, urls []string, title string, mangaTitle string) (string, error) {
	safeManga := sanitize(mangaTitle)
	safeChapter := sanitize(title)

	mangaDir := filepath.Join(d.BaseDir, safeManga)
	if err := os.MkdirAll(mangaDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create manga dir: %w", err)
	}

	cbzPath := filepath.Join(mangaDir, safeChapter+".cbz")
	if _, err := os.Stat(cbzPath); err == nil {
		return cbzPath, fmt.Errorf("chapter already downloaded")
	}

	// Temp dir for images
	tmpDir := filepath.Join(os.TempDir(), "mangaka", "dl_"+chapterID)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	files, err := d.downloadImages(urls, tmpDir)
	if err != nil {
		return "", err
	}

	if err := d.createCBZ(cbzPath, files); err != nil {
		return "", fmt.Errorf("failed to create CBZ: %w", err)
	}

	return cbzPath, nil
}

func (d *Downloader) ReadCBZ(path string) error {
	fmt.Printf("Opening %s...\n", path)
	return opener.Open(path)
}

func (d *Downloader) ListDownloads() (map[string][]string, error) {
	downloads := make(map[string][]string)

	entries, err := os.ReadDir(d.BaseDir)
	if err != nil {
		// If dir doesn't exist, return empty
		if os.IsNotExist(err) {
			return downloads, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			mangaName := entry.Name()
			mangaPath := filepath.Join(d.BaseDir, mangaName)

			files, err := os.ReadDir(mangaPath)
			if err != nil {
				continue
			}

			var chapters []string
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".cbz") {
					chapters = append(chapters, f.Name())
				}
			}
			if len(chapters) > 0 {
				downloads[mangaName] = chapters
			}
		}
	}
	return downloads, nil
}

func (d *Downloader) downloadImages(urls []string, tmpDir string) ([]string, error) {
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
		return nil, <-errChan
	}
	return files, nil
}

func sanitize(s string) string {
	invalid := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	for _, char := range invalid {
		s = strings.ReplaceAll(s, char, "")
	}
	return strings.TrimSpace(s)
}
