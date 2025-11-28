package repository

import (
	"errors"
	"log/slog"

	"github.com/aerscs/theca-public/internal/model"
	customerrors "github.com/aerscs/theca-public/internal/utils/errors"
	"gorm.io/gorm"
)

type Repository interface {
	Register(user *model.User) error
	GetUserByUsername(username string) (*model.User, error)
	GetUserByID(id any) (*model.User, error)
	SaveUser(user *model.User) error
	GetUserByRefreshToken(refreshToken string) (*model.User, error)
	GetUserByEmail(email string) (*model.User, error)

	// Методы для работы с закладками
	AddBookmark(bookmark *model.Bookmark) error
	GetBookmarks(userID uint) ([]model.Bookmark, error)
	GetBookmarkByID(bookmarkID uint) (*model.Bookmark, error)
	UpdateBookmark(bookmark *model.Bookmark) error
	DeleteBookmark(bookmarkID uint) error
}

type repository struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewRepository(db *gorm.DB, log *slog.Logger) Repository {
	return &repository{
		db:  db,
		log: log,
	}
}

func (r *repository) Register(user *model.User) error {
	const op = "repository.Register"
	log := r.log.With(slog.String("op", op), slog.String("username", user.Username))

	tx := r.db.Begin()
	if tx.Error != nil {
		log.Error("failed to begin transaction", slog.String("error", tx.Error.Error()))
		return customerrors.FromGormError(tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var existingUser model.User
	result := tx.Where("email = ? OR username = ?", user.Email, user.Username).
		Select("email", "username").
		First(&existingUser)

	if result.Error == nil {
		tx.Rollback()
		if existingUser.Email == user.Email {
			log.Debug("user with email already exists", slog.String("email", user.Email))
			return customerrors.New(customerrors.CodeUserEmailAlreadyExists, "User with this email already exists")
		}
		if existingUser.Username == user.Username {
			log.Debug("user with username already exists", slog.String("username", user.Username))
			return customerrors.New(customerrors.CodeUserUsernameAlreadyExists, "User with this username already exists")
		}
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		tx.Rollback()
		log.Error("failed to check user uniqueness", slog.String("error", result.Error.Error()))
		return customerrors.FromUserError(result.Error, "users")
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		log.Error("failed to create user", slog.String("error", err.Error()))
		return customerrors.FromUserError(err, "users")
	}

	if err := tx.Commit().Error; err != nil {
		log.Error("failed to commit transaction", slog.String("error", err.Error()))
		return customerrors.FromUserError(err, "users")
	}

	log.Info("user registered successfully", slog.Uint64("user_id", uint64(user.ID)))
	return nil
}

func (r *repository) GetUserByUsername(username string) (*model.User, error) {
	const op = "repository.GetUserByUsername"
	log := r.log.With("op", op)

	var user model.User
	err := r.db.Model(&model.User{}).Where("username = ?", username).First(&user).Error
	if err != nil {
		log.Error("failed to get user by username", "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.New(customerrors.CodeUserNotFound, "User not found")
		}
		return nil, customerrors.FromUserError(err, "users")
	}

	log.Debug("user retrieved by username successfully", "user_id", user.ID, "username", username)
	return &user, nil
}

func (r *repository) GetUserByID(id any) (*model.User, error) {
	const op = "repository.GetUserByID"
	log := r.log.With("op", op)

	var user model.User
	err := r.db.Model(&model.User{}).Where("id = ?", id).First(&user).Error
	if err != nil {
		log.Error("failed to get user by ID", "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.New(customerrors.CodeUserNotFound, "User not found")
		}
		return nil, customerrors.FromUserError(err, "users")
	}

	log.Debug("user retrieved by ID successfully", "user_id", id)
	return &user, nil
}

func (r *repository) SaveUser(user *model.User) error {
	const op = "repository.SaveUser"
	log := r.log.With("op", op)

	err := r.db.Model(&model.User{}).Where("id = ?", user.ID).Save(user).Error
	if err != nil {
		log.Error("failed to save user", "error", err)
		return customerrors.FromUserError(err, "users")
	}

	log.Debug("user saved successfully", "user_id", user.ID)
	return nil
}

func (r *repository) GetUserByRefreshToken(refreshToken string) (*model.User, error) {
	const op = "repository.GetUserByRefreshToken"
	log := r.log.With("op", op)

	var user model.User
	err := r.db.Model(&model.User{}).Where("refresh_token = ?", refreshToken).First(&user).Error
	if err != nil {
		log.Error("failed to get user by refresh token", "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.New(customerrors.CodeInvalidRefreshToken, "Invalid refresh token")
		}
		return nil, customerrors.FromGormError(err)
	}

	log.Debug("user retrieved by refresh token successfully", "user_id", user.ID, "refresh_token", refreshToken)
	return &user, nil
}

func (r *repository) GetUserByEmail(email string) (*model.User, error) {
	const op = "repository.GetUserByEmail"
	log := r.log.With("op", op)

	var user model.User
	err := r.db.Model(&model.User{}).Where("email = ?", email).First(&user).Error
	if err != nil {
		log.Error("failed to get user by email", "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.New(customerrors.CodeUserNotFound, "User not found")
		}
		return nil, customerrors.FromGormError(err)
	}

	log.Debug("user retrieved by email successfully", "user_id", user.ID, "email", email)
	return &user, nil
}

// Реализация методов для работы с закладками

func (r *repository) AddBookmark(bookmark *model.Bookmark) error {
	const op = "repository.AddBookmark"
	log := r.log.With("op", op)

	err := r.db.Create(bookmark).Error
	if err != nil {
		log.Error("failed to create bookmark", "error", err)
		return customerrors.FromGormError(err)
	}

	log.Debug("bookmark created successfully", "bookmark_id", bookmark.ID, "user_id", bookmark.UserID)
	return nil
}

func (r *repository) GetBookmarks(userID uint) ([]model.Bookmark, error) {
	const op = "repository.GetBookmarks"
	log := r.log.With("op", op)

	var bookmarks []model.Bookmark
	err := r.db.Where("user_id = ?", userID).Find(&bookmarks).Error
	if err != nil {
		log.Error("failed to get bookmarks", "error", err, "user_id", userID)
		return nil, customerrors.FromGormError(err)
	}

	log.Debug("bookmarks retrieved successfully", "user_id", userID, "count", len(bookmarks))
	return bookmarks, nil
}

func (r *repository) GetBookmarkByID(bookmarkID uint) (*model.Bookmark, error) {
	const op = "repository.GetBookmarkByID"
	log := r.log.With("op", op)

	var bookmark model.Bookmark
	err := r.db.Where("id = ?", bookmarkID).First(&bookmark).Error
	if err != nil {
		log.Error("failed to get bookmark by ID", "error", err, "bookmark_id", bookmarkID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.New(customerrors.CodeNotFound, "Bookmark not found")
		}
		return nil, customerrors.FromGormError(err)
	}

	log.Debug("bookmark retrieved successfully", "bookmark_id", bookmarkID, "user_id", bookmark.UserID)
	return &bookmark, nil
}

func (r *repository) UpdateBookmark(bookmark *model.Bookmark) error {
	const op = "repository.UpdateBookmark"
	log := r.log.With("op", op)

	err := r.db.Save(bookmark).Error
	if err != nil {
		log.Error("failed to update bookmark", "error", err, "bookmark_id", bookmark.ID)
		return customerrors.FromGormError(err)
	}

	log.Debug("bookmark updated successfully", "bookmark_id", bookmark.ID, "user_id", bookmark.UserID)
	return nil
}

func (r *repository) DeleteBookmark(bookmarkID uint) error {
	const op = "repository.DeleteBookmark"
	log := r.log.With("op", op)

	err := r.db.Delete(&model.Bookmark{}, bookmarkID).Error
	if err != nil {
		log.Error("failed to delete bookmark", "error", err, "bookmark_id", bookmarkID)
		return customerrors.FromGormError(err)
	}

	log.Debug("bookmark deleted successfully", "bookmark_id", bookmarkID)
	return nil
}
