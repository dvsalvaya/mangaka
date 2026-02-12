package service

import (
	"context"
	"encoding/json"
	"fmt"
	"mangaka/internal/api"
	"mangaka/internal/downloader"
	"mangaka/pkg/models"
	"os"
	"sync"
)

type MangaService struct {
	client     *api.Client
	downloader *downloader.Downloader
	favorites  map[string]models.Favorite // Changed Key to string
	favMu      sync.RWMutex
	favFile    string
}

func NewMangaService(client *api.Client, dl *downloader.Downloader) *MangaService {
	s := &MangaService{
		client:     client,
		downloader: dl,
		favorites:  make(map[string]models.Favorite),
		favFile:    "favorites.json",
	}
	s.loadFavorites()
	return s
}

func (s *MangaService) SearchManga(ctx context.Context, query string, offset int) (*models.MangaDexResponse, error) {
	// No cache for now to simplify migration
	return s.client.SearchManga(ctx, query, offset)
}

func (s *MangaService) GetMangaChapters(ctx context.Context, mangaID string) ([]models.Chapter, error) {
	return s.client.GetMangaChapters(ctx, mangaID)
}

func (s *MangaService) ReadChapter(ctx context.Context, chapterID string) error {
	urls, err := s.client.GetChapterPages(ctx, chapterID)
	if err != nil {
		return err
	}
	if len(urls) == 0 {
		return fmt.Errorf("no pages found for chapter")
	}

	return s.downloader.DownloadAndRead(chapterID, urls)
}

func (s *MangaService) ToggleFavorite(manga models.MangaData) bool {
	s.favMu.Lock()
	defer s.favMu.Unlock()

	id := manga.ID
	if _, exists := s.favorites[id]; exists {
		delete(s.favorites, id)
		s.saveFavorites()
		return false // Removed
	}

	title := "Unknown"
	if t, ok := manga.Attributes.Title["en"]; ok {
		title = t
	} else {
		// fallback to first key
		for _, v := range manga.Attributes.Title {
			title = v
			break
		}
	}

	s.favorites[id] = models.Favorite{
		MangaID: id,
		Title:   title,
	}
	s.saveFavorites()
	return true // Added
}

func (s *MangaService) ListFavorites() []models.Favorite {
	s.favMu.RLock()
	defer s.favMu.RUnlock()

	favs := make([]models.Favorite, 0, len(s.favorites))
	for _, f := range s.favorites {
		favs = append(favs, f)
	}
	return favs
}

func (s *MangaService) loadFavorites() {
	file, err := os.Open(s.favFile)
	if err != nil {
		return
	}
	defer file.Close()

	var favList []models.Favorite
	if err := json.NewDecoder(file).Decode(&favList); err == nil {
		s.favMu.Lock()
		for _, f := range favList {
			s.favorites[f.MangaID] = f
		}
		s.favMu.Unlock()
	}
}

func (s *MangaService) saveFavorites() {
	file, err := os.Create(s.favFile)
	if err != nil {
		fmt.Printf("Error saving favorites: %v\n", err)
		return
	}
	defer file.Close()

	favs := make([]models.Favorite, 0, len(s.favorites))
	for _, f := range s.favorites {
		favs = append(favs, f)
	}

	json.NewEncoder(file).Encode(favs)
}
