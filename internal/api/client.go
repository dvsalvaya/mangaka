package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mangaka/pkg/models"
	"net/http"
	"net/url"
	"time"
)

var ErrNotFound = errors.New("resource not found")

const BaseURL = "https://api.mangadex.org"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) SearchManga(ctx context.Context, query string, offset int) (*models.MangaDexResponse, error) {
	limit := 10
	endpoint := fmt.Sprintf("%s/manga?title=%s&limit=%d&offset=%d&contentRating[]=safe&contentRating[]=suggestive", BaseURL, url.QueryEscape(query), limit, offset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.MangaDexResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding error: %w", err)
	}

	return &result, nil
}

func (c *Client) GetMangaChapters(ctx context.Context, mangaID string) ([]models.Chapter, error) {
	endpoint := fmt.Sprintf("%s/manga/%s/feed?translatedLanguage[]=en&limit=96&order[chapter]=desc", BaseURL, mangaID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.MangaDexChapterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var chapters []models.Chapter
	for _, data := range result.Data {
		title := data.Attributes.Title
		chNum := data.Attributes.Chapter
		extURL := data.Attributes.ExternalURL // Map external URL

		displayTitle := fmt.Sprintf("Ch. %s", chNum)
		if title != "" {
			displayTitle = fmt.Sprintf("%s - %s", displayTitle, title)
		}

		chapters = append(chapters, models.Chapter{
			ID:          data.ID,
			Title:       displayTitle,
			Chapter:     chNum,
			Pages:       data.Attributes.Pages,
			ExternalURL: extURL,
		})
	}

	return chapters, nil
}

func (c *Client) GetChapterPages(ctx context.Context, chapterID string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/at-home/server/%s", BaseURL, chapterID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.MangaDexAtHomeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var urls []string
	for _, filename := range result.Chapter.Data {
		u := fmt.Sprintf("%s/data/%s/%s", result.BaseURL, result.Chapter.Hash, filename)
		urls = append(urls, u)
	}

	return urls, nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "Mangaka-CLI/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	return resp, nil
}
