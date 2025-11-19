package model

import "time"

// Bookmark представляет собой модель закладки
type Bookmark struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Favicon   string    `json:"favicon"`
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	ShowText  bool      `json:"show_text"`
}

// ImportBookmarksRequest представляет запрос на импорт закладок
type ImportBookmarksRequest struct {
	File string `json:"file" binding:"required"`
}

// ExportBookmarksResponse представляет ответ на экспорт закладок
type ExportBookmarksResponse struct {
	File string `json:"file"`
}


type BookmarkV2Request struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Favicon   string    `json:"favicon"`
	ID        uint      `json:"-"`
	UserID    uint      `json:"user_id"`
	ShowText  bool      `json:"show_text"`
}

type ImportBookmarksV2Request struct {
	Bookmarks []BookmarkV2Request `json:"bookmarks"`

}
