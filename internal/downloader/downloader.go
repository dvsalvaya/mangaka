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
	fmt.Printf("Opening with nsxiv...\n")

    // Opção A: Passar a pasta inteira (O nsxiv ordena alfabeticamente: 001.jpg, 002.jpg...)
    // O "-t" abre no modo galeria (thumbnail). Remova se quiser ver a imagem direto.
    cmd := exec.Command("nsxiv","-f","-b","-s","w", tmpDir)
    
    // IMPORTANTE: Conectar o terminal para os atalhos de teclado do nsxiv funcionarem
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // cmd.Start() roda em background e o programa Go pode fechar e apagar a pasta temporária
    // enquanto ainda está lendo. O cmd.Run() faz o Go esperar você fechar a janela.
    return cmd.Run()
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
