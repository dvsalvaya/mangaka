package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/dvsalvaya/mangaka/internal/api"
	"github.com/dvsalvaya/mangaka/internal/downloader"
	"github.com/dvsalvaya/mangaka/pkg/models"
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

func (s *MangaService) DownloadChapter(ctx context.Context, chapter models.Chapter, manga models.MangaData) (string, error) {
	urls, err := s.client.GetChapterPages(ctx, chapter.ID)
	if err != nil {
		return "", err
	}
	if len(urls) == 0 {
		return "", fmt.Errorf("no pages found for chapter")
	}

	mangaTitle := manga.Attributes.Title["en"]
	if mangaTitle == "" {
		for _, t := range manga.Attributes.Title {
			mangaTitle = t
			break
		}
	}
	if mangaTitle == "" {
		mangaTitle = "Unknown Manga"
	}

	return s.downloader.DownloadChapter(chapter.ID, urls, chapter.Title, mangaTitle)
}

func (s *MangaService) ListDownloads() (map[string][]string, error) {
	return s.downloader.ListDownloads()
}

func (s *MangaService) ReadDownloaded(mangaTitle, chapterFile string) error {
	// Construct path or use downloader to find it
	// Since ListDownloads returns filenames, and we know BaseDir structure
	// We can reconstruct path: BaseDir/MangaTitle/ChapterFile
	// But Service shouldn't know about BaseDir path details ideally.
	// Either modify ListDownloads to return full paths, or add ReadLocal(manga, chapter) to downloader.
	// Let's use filepath.Join here assuming standard structure or add a method in downloader.
	// For now, let's construct it here as we don't have GetPath in downloader interface yet.
	// Wait, downloader.ReadCBZ takes a path.
	// I'll add a helper in service to build path, or just use relative path?
	// The downloader is in internal/downloader, service in internal/service.
	// I can use `downloader.DownloadDir` constant if exported? It was `const DownloadDir = "downloads"`.
	// I didn't export it (captital D). I did `const DownloadDir`. Yes, it is exported.
	path := fmt.Sprintf("%s/%s/%s", downloader.DownloadDir, mangaTitle, chapterFile)
	return s.downloader.ReadCBZ(path)
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
