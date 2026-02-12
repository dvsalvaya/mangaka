package main

import (
	"mangaka/internal/api"
	"mangaka/internal/downloader"
	"mangaka/internal/service"
	"mangaka/internal/ui"
)

func main() {
	client := api.NewClient()
	dl := downloader.NewDownloader()
	svc := service.NewMangaService(client, dl)
	cli := ui.NewCLI(svc)
	cli.Start()
}
