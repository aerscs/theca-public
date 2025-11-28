package parsers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/aerscs/theca-public/internal/model"
)

// BookmarkHTMLExporter структура для экспорта закладок в HTML-файл
type BookmarkHTMLExporter struct{}

// NewBookmarkHTMLExporter создает новый экземпляр экспортера закладок
func NewBookmarkHTMLExporter() *BookmarkHTMLExporter {
	return &BookmarkHTMLExporter{}
}

// ExportToHTML экспортирует закладки в HTML-формат
func (e *BookmarkHTMLExporter) ExportToHTML(bookmarks []model.Bookmark) (string, error) {
	var buffer bytes.Buffer

	// HTML header
	buffer.WriteString(`<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<meta http-equiv="Content-Security-Policy"
      content="default-src 'self'; script-src 'none'; img-src data: *; object-src 'none'"></meta>
<TITLE>Bookmarks</TITLE>
<H1>Меню закладок</H1>

<DL><p>
`)

	// Добавление категории "Панель закладок"
	buffer.WriteString(fmt.Sprintf(`<DT><H3 ADD_DATE="%d" LAST_MODIFIED="%d" PERSONAL_TOOLBAR_FOLDER="true">Панель закладок</H3>
<DL><p>
`, time.Now().Unix(), time.Now().Unix()))

	// Добавление закладок
	for _, bookmark := range bookmarks {
		addDate := bookmark.CreatedAt.Unix()
		lastModified := bookmark.UpdatedAt.Unix()
		title := sanitizeHTML(bookmark.Title)
		url := bookmark.URL
		favicon := bookmark.Favicon

		// Если название пустое, используем URL
		if title == "" {
			title = url
		}

		buffer.WriteString(fmt.Sprintf(`<DT><A HREF="%s" ADD_DATE="%d" LAST_MODIFIED="%d" ICON_URI="%s">%s</A>
`, url, addDate, lastModified, favicon, title))
	}

	// Закрытие HTML
	buffer.WriteString(`</DL><p>
</DL>
`)

	// Кодируем результат в base64
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

// sanitizeHTML экранирует специальные символы HTML
func sanitizeHTML(input string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(input)
}

// ExportBookmarksToHTML обертка для удобного экспорта закладок
func ExportBookmarksToHTML(bookmarks []model.Bookmark) (string, error) {
	exporter := NewBookmarkHTMLExporter()
	return exporter.ExportToHTML(bookmarks)
}
