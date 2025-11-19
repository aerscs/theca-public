package handlers

import (
	"strconv"

	"github.com/OxytocinGroup/theca-v3/internal/model"
	"github.com/OxytocinGroup/theca-v3/internal/utils/errors"
	"github.com/gin-gonic/gin"
)

// @Summary Add Bookmark
// @Description Add a new bookmark
// @Tags bookmarks
// @Accept json
// @Produce json
// @Param bookmarkRequest body model.AddBookmarkRequest true "Bookmark data"
// @Success 200 {object} model.BookmarkResponse
// @Failure 400
// @Failure 401
// @Failure 500
// @Security Bearer
// @Router /v1/api/bookmarks [post]
func (h *Handler) AddBookmark(c *gin.Context) {
	const op = "handler.AddBookmark"
	log := h.log.With("op", op)

	userID := c.GetUint("userID")
	if userID == 0 {
		log.Error("user ID not found in context")
		errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Unauthorized"))
		return
	}

	var req model.AddBookmarkRequest
	if err := c.BindJSON(&req); err != nil {
		log.Debug("binding json", "err", err, "req", req)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "Invalid request format"))
		return
	}

	bookmark, err := h.service.AddBookmark(userID, req.Title, req.URL, req.ShowText)
	if err != nil {
		log.Error("failed to add bookmark", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("bookmark added successfully", "user_id", userID, "bookmark_id", bookmark.ID)
	errors.RespondWithSuccess(c, model.BookmarkResponse{
		ID:        bookmark.ID,
		Title:     bookmark.Title,
		URL:       bookmark.URL,
		ShowText:  bookmark.ShowText,
		CreatedAt: bookmark.CreatedAt,
		UpdatedAt: bookmark.UpdatedAt,
		Favicon:   bookmark.Favicon,
	})
}

// @Summary Get All Bookmarks
// @Description Get all bookmarks for the authenticated user
// @Tags bookmarks
// @Produce json
// @Success 200 {array} model.BookmarkResponse
// @Failure 401
// @Failure 500
// @Security Bearer
// @Router /v1/api/bookmarks [get]
func (h *Handler) GetBookmarks(c *gin.Context) {
	const op = "handler.GetBookmarks"
	log := h.log.With("op", op)

	userID := c.GetUint("userID")
	if userID == 0 {
		log.Error("user ID not found in context")
		errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Unauthorized"))
		return
	}

	bookmarks, err := h.service.GetBookmarks(userID)
	if err != nil {
		log.Error("failed to get bookmarks", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	bookmarkResponses := make([]model.BookmarkResponse, len(bookmarks))
	for i, bookmark := range bookmarks {
		bookmarkResponses[i] = model.BookmarkResponse{
			ID:        bookmark.ID,
			Title:     bookmark.Title,
			URL:       bookmark.URL,
			ShowText:  bookmark.ShowText,
			CreatedAt: bookmark.CreatedAt,
			UpdatedAt: bookmark.UpdatedAt,
			Favicon:   bookmark.Favicon,
		}
	}

	log.Debug("bookmarks retrieved successfully", "user_id", userID, "count", len(bookmarks))
	errors.RespondWithSuccess(c, bookmarkResponses)
}

// @Summary Get Bookmark By ID
// @Description Get a bookmark by its ID
// @Tags bookmarks
// @Produce json
// @Param id path int true "Bookmark ID"
// @Success 200 {object} model.BookmarkResponse
// @Failure 400
// @Failure 401
// @Failure 403
// @Failure 404
// @Failure 500
// @Security Bearer
// @Router /v1/api/bookmarks/{id} [get]
func (h *Handler) GetBookmarkByID(c *gin.Context) {
	const op = "handler.GetBookmarkByID"
	log := h.log.With("op", op)

	userID := c.GetUint("userID")
	if userID == 0 {
		log.Error("user ID not found in context")
		errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Unauthorized"))
		return
	}

	bookmarkIDStr := c.Param("id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		log.Error("invalid bookmark ID", "error", err, "bookmark_id", bookmarkIDStr)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "Invalid bookmark ID"))
		return
	}

	bookmark, err := h.service.GetBookmarkByID(userID, uint(bookmarkID))
	if err != nil {
		log.Error("failed to get bookmark", "error", err, "bookmark_id", bookmarkID)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("bookmark retrieved successfully", "user_id", userID, "bookmark_id", bookmarkID)
	errors.RespondWithSuccess(c, model.BookmarkResponse{
		ID:        bookmark.ID,
		Title:     bookmark.Title,
		URL:       bookmark.URL,
		ShowText:  bookmark.ShowText,
		CreatedAt: bookmark.CreatedAt,
		UpdatedAt: bookmark.UpdatedAt,
		Favicon:   bookmark.Favicon,
	})
}

// @Summary Update Bookmark
// @Description Update an existing bookmark
// @Tags bookmarks
// @Accept json
// @Produce json
// @Param id path int true "Bookmark ID"
// @Param bookmarkRequest body model.PatchBookmarkRequest true "Partial bookmark data for update"
// @Success 200 {object} model.BookmarkResponse
// @Failure 400
// @Failure 401
// @Failure 403
// @Failure 404
// @Failure 500
// @Security Bearer
// @Router /v1/api/bookmarks/{id} [patch]
func (h *Handler) UpdateBookmark(c *gin.Context) {
	const op = "handler.UpdateBookmark"
	log := h.log.With("op", op)

	userID := c.GetUint("userID")
	if userID == 0 {
		log.Error("user ID not found in context")
		errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Unauthorized"))
		return
	}

	bookmarkIDStr := c.Param("id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		log.Error("invalid bookmark ID", "error", err, "bookmark_id", bookmarkIDStr)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "Invalid bookmark ID"))
		return
	}

	var req model.PatchBookmarkRequest
	if err := c.BindJSON(&req); err != nil {
		log.Debug("binding json", "err", err, "req", req)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "Invalid request format"))
		return
	}

	if req.Title == nil && req.URL == nil && req.ShowText == nil {
		log.Debug("empty patch request")
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "No fields to update"))
		return
	}

	bookmark, err := h.service.PatchBookmark(userID, uint(bookmarkID), &req)
	if err != nil {
		log.Error("failed to update bookmark", "error", err, "bookmark_id", bookmarkID)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("bookmark updated successfully", "user_id", userID, "bookmark_id", bookmarkID)
	errors.RespondWithSuccess(c, model.BookmarkResponse{
		ID:        bookmark.ID,
		Title:     bookmark.Title,
		URL:       bookmark.URL,
		ShowText:  bookmark.ShowText,
		CreatedAt: bookmark.CreatedAt,
		UpdatedAt: bookmark.UpdatedAt,
		Favicon:   bookmark.Favicon,
	})
}

// @Summary Delete Bookmark
// @Description Delete a bookmark
// @Tags bookmarks
// @Produce json
// @Param id path int true "Bookmark ID"
// @Success 200 {object} errors.Response
// @Failure 400
// @Failure 401
// @Failure 403
// @Failure 404
// @Failure 500
// @Security Bearer
// @Router /v1/api/bookmarks/{id} [delete]
func (h *Handler) DeleteBookmark(c *gin.Context) {
	const op = "handler.DeleteBookmark"
	log := h.log.With("op", op)

	userID := c.GetUint("userID")
	if userID == 0 {
		log.Error("user ID not found in context")
		errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Unauthorized"))
		return
	}

	bookmarkIDStr := c.Param("id")
	bookmarkID, err := strconv.ParseUint(bookmarkIDStr, 10, 32)
	if err != nil {
		log.Error("invalid bookmark ID", "error", err, "bookmark_id", bookmarkIDStr)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "Invalid bookmark ID"))
		return
	}

	err = h.service.DeleteBookmark(userID, uint(bookmarkID))
	if err != nil {
		log.Error("failed to delete bookmark", "error", err, "bookmark_id", bookmarkID)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("bookmark deleted successfully", "user_id", userID, "bookmark_id", bookmarkID)
	errors.RespondWithSuccess(c, "Bookmark deleted successfully")
}

// @Summary Import Bookmarks
// @Description Import bookmarks from HTML file encoded in base64
// @Tags bookmarks
// @Accept json
// @Produce json
// @Param importRequest body model.ImportBookmarksRequest true "Import data"
// @Success 200 {array} model.BookmarkResponse
// @Failure 400
// @Failure 401
// @Failure 500
// @Security Bearer
// @Router /v1/api/bookmarks/import [put]
func (h *Handler) ImportBookmarks(c *gin.Context) {
	const op = "handler.ImportBookmarks"
	log := h.log.With("op", op)

	userID := c.GetUint("userID")
	if userID == 0 {
		log.Error("user ID not found in context")
		errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Unauthorized"))
		return
	}

	var req model.ImportBookmarksRequest
	if err := c.BindJSON(&req); err != nil {
		log.Debug("binding json", "err", err)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "Invalid request format"))
		return
	}

	if req.File == "" {
		log.Debug("empty file data")
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "File data is required"))
		return
	}

	bookmarks, err := h.service.ImportBookmarks(userID, req.File)
	if err != nil {
		log.Error("failed to import bookmarks", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	bookmarkResponses := make([]model.BookmarkResponse, len(bookmarks))
	for i, bookmark := range bookmarks {
		bookmarkResponses[i] = model.BookmarkResponse{
			ID:        bookmark.ID,
			Title:     bookmark.Title,
			URL:       bookmark.URL,
			ShowText:  bookmark.ShowText,
			CreatedAt: bookmark.CreatedAt,
			UpdatedAt: bookmark.UpdatedAt,
			Favicon:   bookmark.Favicon,
		}
	}

	log.Debug("bookmarks imported successfully", "user_id", userID, "count", len(bookmarks))
	errors.RespondWithSuccess(c, bookmarkResponses)
}

// @Summary Export Bookmarks
// @Description Export all user's bookmarks as HTML file in base64 encoding
// @Tags bookmarks
// @Produce json
// @Success 200 {object} model.ExportBookmarksResponse
// @Failure 401
// @Failure 500
// @Security Bearer
// @Router /v1/api/bookmarks/export [get]
func (h *Handler) ExportBookmarks(c *gin.Context) {
	const op = "handler.ExportBookmarks"
	log := h.log.With("op", op)

	userID := c.GetUint("userID")
	if userID == 0 {
		log.Error("user ID not found in context")
		errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Unauthorized"))
		return
	}

	base64Data, err := h.service.ExportBookmarks(userID)
	if err != nil {
		log.Error("failed to export bookmarks", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("bookmarks exported successfully", "user_id", userID)
	errors.RespondWithSuccess(c, model.ExportBookmarksResponse{
		File: base64Data,
	})
}

// @Summary Import Bookmarks V2
// @Description Import bookmarks from HTML file encoded in base64
// @Tags bookmarks
// @Accept json
// @Produce json
// @Param importRequest body model.ImportBookmarksRequest true "Import data"
// @Success 200 {array} model.BookmarkResponse
// @Failure 400
// @Failure 401
// @Failure 500
// @Security Bearer
// @Router /v2/api/bookmarks/import [POST]
func (h *Handler) ImportBookmarksV2(c *gin.Context) {
	const op = "handler.ImportBookmarksV2"
	log := h.log.With("op", op)

	userID := c.GetUint("userID")
	if userID == 0 {
		log.Error("user ID not found in context")
		errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Unauthorized"))
		return
	}

	var req model.ImportBookmarksV2Request
	if err := c.BindJSON(&req); err != nil {
		log.Debug("binding json", "err", err)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "Invalid request format"))
		return
	}

	bookmarks, err := h.service.ImportBookmarksV2(userID, req.Bookmarks)
	if err != nil {
		log.Error("failed to import bookmarks", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("bookmarks imported successfully", "user_id", userID, "count", len(bookmarks))
	errors.RespondWithSuccess(c, bookmarks)
}

// @Summary Export Bookmarks V2
// @Description Export all user's bookmarks as JSON
// @Tags bookmarks
// @Produce json
// @Success 200 {array} model.BookmarkResponse
// @Failure 401
// @Failure 500
// @Security Bearer
// @Router /v2/api/bookmarks/export [get]
func (h *Handler) ExportBookmarksV2(c *gin.Context) {
	const op = "handler.ExportBookmarksV2"
	log := h.log.With("op", op)

	userID := c.GetUint("userID")
	if userID == 0 {
		log.Error("user ID not found in context")
		errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Unauthorized"))
		return
	}

	bookmarks, err := h.service.ExportBookmarksV2(userID)
	if err != nil {
		log.Error("failed to export bookmarks", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	c.Writer.Header().Set("Content-Disposition", "attachment; filename=bookmarks.json")
	c.Writer.Header().Set("Content-Type", "application/json")
	log.Debug("bookmarks exported successfully", "user_id", userID, "count", len(bookmarks))
	errors.RespondWithSuccess(c, bookmarks)
}
