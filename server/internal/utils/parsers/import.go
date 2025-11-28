package parsers

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/url"
	"sync"

	"github.com/aerscs/theca-public/internal/model"
	"github.com/aerscs/theca-public/internal/repository"
	"golang.org/x/net/html"
)

// BookmarkHTMLParser структура для парсинга HTML-файла закладок
type BookmarkHTMLParser struct {
	faviconCache repository.FaviconCacheRepository
}

// NewBookmarkHTMLParser создает новый экземпляр парсера закладок
func NewBookmarkHTMLParser(faviconCache repository.FaviconCacheRepository) *BookmarkHTMLParser {
	return &BookmarkHTMLParser{
		faviconCache: faviconCache,
	}
}

// ParseHTML парсит HTML-файл закладок, закодированный в base64
func (p *BookmarkHTMLParser) ParseHTML(ctx context.Context, base64Data string) ([]model.Bookmark, error) {
	// decode base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, err
	}

	// parse html
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// extract bookmarks (без фавиконок)
	bookmarks := make([]model.Bookmark, 0)
	p.traverseHTML(ctx, doc, &bookmarks)

	// параллельно получаем фавиконки
	p.fetchFaviconsParallel(ctx, bookmarks)

	return bookmarks, nil
}

// getFavicon получает favicon по URL закладки в формате base64
func (p *BookmarkHTMLParser) getFavicon(ctx context.Context, bookmarkURL string) string {
	if bookmarkURL == "" {
		return ""
	}

	favicon, err := FetchFaviconBase64(ctx, p.faviconCache, bookmarkURL)
	if err != nil {
		return ""
	}

	return favicon
}

// fetchFaviconsParallel параллельно получает фавиконки для всех закладок
func (p *BookmarkHTMLParser) fetchFaviconsParallel(ctx context.Context, bookmarks []model.Bookmark) {
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, 10)

	for i := range bookmarks {
		if bookmarks[i].URL == "" {
			continue
		}

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			favicon := p.getFavicon(ctx, bookmarks[idx].URL)

			bookmarks[idx].Favicon = favicon
		}(i)
	}

	wg.Wait()
}

// traverseHTML рекурсивно обходит HTML-дерево и извлекает закладки
func (p *BookmarkHTMLParser) traverseHTML(ctx context.Context, n *html.Node, bookmarks *[]model.Bookmark) {
	if n.Type == html.ElementNode && n.Data == "a" {
		// this is a bookmark (tag <a>)
		var bookmarkURL, title string

		// extract URL
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				bookmarkURL = attr.Val
				break
			}
		}

		// check if URL is valid
		if bookmarkURL != "" {
			_, err := url.Parse(bookmarkURL)
			if err != nil {
				// skip invalid URLs
				goto nextNode
			}

			// extract bookmark text (title)
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				title = n.FirstChild.Data
			}

			// создаем закладку без фавиконки
			*bookmarks = append(*bookmarks, model.Bookmark{
				Title: title,
				URL:   bookmarkURL,
			})
		}
	}

nextNode:
	// recursively traverse all child elements
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.traverseHTML(ctx, c, bookmarks)
	}
}

// ParseBookmarksFromHTML wrapper for convenient bookmarks import
func ParseBookmarksFromHTML(ctx context.Context, base64Data string, faviconCache repository.FaviconCacheRepository) ([]model.Bookmark, error) {
	parser := NewBookmarkHTMLParser(faviconCache)
	return parser.ParseHTML(ctx, base64Data)
}
