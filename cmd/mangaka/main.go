package main

import (
	"github.com/dvsalvaya/mangaka/internal/api"
	"github.com/dvsalvaya/mangaka/internal/downloader"
	"github.com/dvsalvaya/mangaka/internal/service"
	"github.com/dvsalvaya/mangaka/internal/ui"
)

func main() {
	client := api.NewClient()
	dl := downloader.NewDownloader()
	svc := service.NewMangaService(client, dl)
	cli := ui.NewCLI(svc)
	cli.Start()
}
