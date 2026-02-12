package ui

import (
	"context"
	"fmt"

	"github.com/dvsalvaya/mangaka/internal/service"
	"github.com/dvsalvaya/mangaka/pkg/models"

	"github.com/manifoldco/promptui"
)

type CLI struct {
	service *service.MangaService
}

func NewCLI(service *service.MangaService) *CLI {
	return &CLI{
		service: service,
	}
}

func (c *CLI) Start() {
	fmt.Println("Welcome to Mangaka CLI! (MangaDex + Zathura Edition)")
	for {
		prompt := promptui.Select{
			Label: "Main Menu",
			Items: []string{"Search Manga", "My Favorites", "My Downloads", "Exit"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Select failed %v\n", err)
			return
		}

		switch result {
		case "Search Manga":
			c.searchFlow()
		case "My Favorites":
			c.favoritesFlow()
		case "My Downloads":
			c.downloadsFlow()
		case "Exit":
			fmt.Println("Goodbye!")
			return
		}
	}
}

func (c *CLI) searchFlow() {
	prompt := promptui.Prompt{
		Label: "Search for Manga",
	}

	query, err := prompt.Run()
	if err != nil {
		return
	}

	offset := 0
	for {
		result, err := c.service.SearchManga(context.Background(), query, offset)
		if err != nil {
			fmt.Printf("Error searching: %v\n", err)
			return
		}

		if len(result.Data) == 0 {
			fmt.Println("No results found.")
			return
		}

		items := make([]string, len(result.Data))
		for i, m := range result.Data {
			title := m.Attributes.Title["en"]
			if title == "" {
				// fallback
				for _, t := range m.Attributes.Title {
					title = t
					break
				}
			}
			items[i] = title
		}

		if result.Total > offset+result.Limit {
			items = append(items, "Next Page >>")
		}
		items = append(items, "<< Back to Main Menu")

		selectPrompt := promptui.Select{
			Label: fmt.Sprintf("Results for '%s'", query),
			Items: items,
			Size:  10,
		}

		idx, selection, err := selectPrompt.Run()
		if err != nil {
			return
		}

		if selection == "<< Back to Main Menu" {
			return
		}

		if selection == "Next Page >>" {
			offset += result.Limit
			continue
		}

		selectedManga := result.Data[idx]
		c.mangaDetailsFlow(selectedManga)
	}
}

func (c *CLI) mangaDetailsFlow(manga models.MangaData) {
	for {
		title := manga.Attributes.Title["en"]
		if title == "" {
			for _, t := range manga.Attributes.Title {
				title = t
				break
			}
		}

		isFav := false
		for _, f := range c.service.ListFavorites() {
			if f.MangaID == manga.ID {
				isFav = true
				break
			}
		}

		favLabel := "Add to Favorites"
		if isFav {
			favLabel = "Remove from Favorites"
		}

		prompt := promptui.Select{
			Label: fmt.Sprintf("Manga: %s", title),
			Items: []string{"List Chapters", favLabel, "<< Back to Results"},
		}

		_, result, err := prompt.Run()
		if err != nil {
			return
		}

		switch result {
		case "List Chapters":
			c.chaptersFlow(manga)
		case favLabel:
			added := c.service.ToggleFavorite(manga)
			if added {
				fmt.Println("Added to favorites.")
			} else {
				fmt.Println("Removed from favorites.")
			}
		case "<< Back to Results":
			return
		}
	}
}

func (c *CLI) chaptersFlow(manga models.MangaData) {
	fmt.Println("Fetching chapters...")
	chapters, err := c.service.GetMangaChapters(context.Background(), manga.ID)
	if err != nil {
		fmt.Printf("Error getting chapters: %v\n", err)
		return
	}

	if len(chapters) == 0 {
		fmt.Println("No chapters found.")
		return
	}

	items := make([]string, len(chapters))
	for i, ch := range chapters {
		title := ch.Title
		if ch.ExternalURL != "" {
			title = fmt.Sprintf("%s [External]", title)
		}
		items[i] = title
	}
	items = append(items, "<< Back")

	prompt := promptui.Select{
		Label: "Select Chapter to Read",
		Items: items,
		Size:  15,
	}

	idx, result, err := prompt.Run()
	if err != nil || result == "<< Back" {
		return
	}

	selectedChapter := chapters[idx]

	if selectedChapter.ExternalURL != "" {
		fmt.Printf("This chapter is hosted externally. Please open in browser:\n%s\n", selectedChapter.ExternalURL)
		c.waitForKey()
		return
	}

	actionPrompt := promptui.Select{
		Label: fmt.Sprintf("Chapter: %s", selectedChapter.Title),
		Items: []string{"Read Online", "Download", "Cancel"},
	}

	_, action, err := actionPrompt.Run()
	if err != nil || action == "Cancel" {
		return
	}

	if action == "Download" {
		fmt.Printf("Downloading '%s'...\n", selectedChapter.Title)
		path, err := c.service.DownloadChapter(context.Background(), selectedChapter, manga)
		if err != nil {
			fmt.Printf("Error downloading: %v\n", err)
		} else {
			fmt.Printf("Downloaded to %s\n", path)
		}
		c.waitForKey()
		return
	}

	// Read Online
	fmt.Printf("Downloading and opening '%s'...\n", selectedChapter.Title)

	err = c.service.ReadChapter(context.Background(), selectedChapter.ID)
	if err != nil {
		fmt.Printf("Error reading chapter: %v\n", err)
	}
}

func (c *CLI) downloadsFlow() {
	downloads, err := c.service.ListDownloads()
	if err != nil {
		fmt.Printf("Error listing downloads: %v\n", err)
		return
	}

	if len(downloads) == 0 {
		fmt.Println("No downloads found.")
		return
	}

	mangaTitles := make([]string, 0, len(downloads))
	for title := range downloads {
		mangaTitles = append(mangaTitles, title)
	}
	mangaTitles = append(mangaTitles, "<< Back to Main Menu")

	prompt := promptui.Select{
		Label: "My Downloads",
		Items: mangaTitles,
	}

	idx, result, err := prompt.Run()
	if err != nil || result == "<< Back to Main Menu" {
		return
	}

	selectedManga := mangaTitles[idx]
	chapters := downloads[selectedManga]

	// Add Back option
	chapters = append(chapters, "<< Back")

	chapPrompt := promptui.Select{
		Label: fmt.Sprintf("Downloads for %s", selectedManga),
		Items: chapters,
	}

	_, cResult, err := chapPrompt.Run()
	if err != nil || cResult == "<< Back" {
		return
	}

	// Read
	err = c.service.ReadDownloaded(selectedManga, cResult)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		c.waitForKey()
	}
}

func (c *CLI) favoritesFlow() {
	favs := c.service.ListFavorites()
	if len(favs) == 0 {
		fmt.Println("No favorites yet.")
		return
	}

	items := make([]string, len(favs))
	for i, f := range favs {
		items[i] = f.Title
	}
	items = append(items, "<< Back to Main Menu")

	prompt := promptui.Select{
		Label: "My Favorites",
		Items: items,
	}

	idx, result, err := prompt.Run()
	if err != nil || result == "<< Back to Main Menu" {
		return
	}

	selectedFav := favs[idx]
	manga := models.MangaData{
		ID: selectedFav.MangaID,
		Attributes: models.MangaAttributes{
			Title: map[string]string{"en": selectedFav.Title},
		},
	}

	c.mangaDetailsFlow(manga)
}

func (c *CLI) waitForKey() {
	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}
