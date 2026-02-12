package models

type Manga struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Year        int    `json:"year"`
}

type Chapter struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Chapter     string `json:"chapter"`
	Pages       int    `json:"pages"`
	ExternalURL string `json:"externalUrl"` // New field
}

// API Responses

type MangaDexResponse struct {
	Data   []MangaData `json:"data"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	Total  int         `json:"total"`
}

type MangaData struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"` // "manga"
	Attributes MangaAttributes `json:"attributes"`
}

type MangaAttributes struct {
	Title       map[string]string `json:"title"`
	Description map[string]string `json:"description"`
	Status      string            `json:"status"`
	Year        int               `json:"year"`
}

type MangaDexChapterResponse struct {
	Data   []ChapterData `json:"data"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

type ChapterData struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"` // "chapter"
	Attributes ChapterAttributes `json:"attributes"`
}

type ChapterAttributes struct {
	Title       string `json:"title"`
	Chapter     string `json:"chapter"`
	Pages       int    `json:"pages"`
	PublishAt   string `json:"publishAt"`
	ExternalURL string `json:"externalUrl"` // New field in attributes
}

type MangaDexAtHomeResponse struct {
	BaseURL string `json:"baseUrl"`
	Chapter struct {
		Hash string   `json:"hash"`
		Data []string `json:"data"`
	} `json:"chapter"`
}

type Favorite struct {
	MangaID string `json:"manga_id"`
	Title   string `json:"title"`
}
