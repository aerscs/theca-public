package service

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/aerscs/theca-public/internal/config"
	"github.com/aerscs/theca-public/internal/model"
	"github.com/aerscs/theca-public/internal/repository"
	"github.com/aerscs/theca-public/internal/utils/errors"
	jwtauth "github.com/aerscs/theca-public/internal/utils/jwt"
	"github.com/aerscs/theca-public/internal/utils/mail"
	"github.com/aerscs/theca-public/internal/utils/parsers"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(req *model.RegisterRequest) (uint, error)
	Login(username, password string) (string, string, *model.User, error)
	LogoutFromAllSessions(userID uint) error
	VerifyEmail(code string) (string, string, *model.User, error)
	SendEmailVerificationCode(email string) error
	RefreshTokens(refreshToken string) (string, string, error)
	RequestPasswordReset(email string) error
	ResetPassword(token, password string) error
	GetUser(userID any) (*model.UserResponse, error)

	// Методы для работы с закладками
	AddBookmark(userID uint, title, url string, showText bool) (*model.Bookmark, error)
	GetBookmarks(userID uint) ([]model.Bookmark, error)
	GetBookmarkByID(userID, bookmarkID uint) (*model.Bookmark, error)
	PatchBookmark(userID, bookmarkID uint, patch *model.PatchBookmarkRequest) (*model.Bookmark, error)
	DeleteBookmark(userID, bookmarkID uint) error
	ImportBookmarks(userID uint, base64Data string) ([]model.Bookmark, error)
	ExportBookmarks(userID uint) (string, error)
	ImportBookmarksV2(userID uint, bookmarks []model.BookmarkV2Request) ([]model.Bookmark, error)
	ExportBookmarksV2(userID uint) ([]model.Bookmark, error)
}

type service struct {
	repo   repository.Repository
	cache  repository.CacheRepository
	log    *slog.Logger
	cfg    *config.Config
	mailer mail.Mailer
}

func NewService(repo repository.Repository, cache repository.CacheRepository, log *slog.Logger, cfg *config.Config) Service {
	return &service{
		repo:   repo,
		cache:  cache,
		log:    log,
		cfg:    cfg,
		mailer: mail.NewMailer(cfg),
	}
}

func (s *service) Register(req *model.RegisterRequest) (uint, error) {
	const op = "service.Register"
	log := s.log.With(slog.String("op", op), slog.String("username", req.Username))

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", slog.String("error", err.Error()))
		return 0, errors.New(errors.CodeInternalError, "failed to process password")
	}

	verificationCode, err := s.generateSecureVerificationCode()
	if err != nil {
		log.Error("failed to generate verification code", slog.String("error", err.Error()))
		return 0, errors.New(errors.CodeInternalError, "failed to generate verification code")
	}

	user := &model.User{
		Email:      req.Email,
		Username:   req.Username,
		PassHash:   string(hashPassword),
		IsVerified: false,
		IsPremium:  false,
	}

	if err := s.repo.Register(user); err != nil {
		log.Error("failed to register user in repository", slog.String("error", err.Error()))
		return 0, err
	}

	// Сохраняем код верификации в Redis
	ctx := context.Background()
	if err := s.cache.StoreEmailVerificationCode(ctx, user.ID, verificationCode); err != nil {
		log.Error("failed to store verification code in Redis", slog.String("error", err.Error()))
	}

	go s.sendVerificationEmailAsync(req.Email, verificationCode, req.Username, user.ID)

	log.Info("user registered successfully", slog.Uint64("user_id", uint64(user.ID)))
	return user.ID, nil
}

func (s *service) generateSecureVerificationCode() (string, error) {
	bytes := make([]byte, 3)
	if _, err := cryptorand.Read(bytes); err != nil {
		return "", err
	}

	code := int(bytes[0])<<16 | int(bytes[1])<<8 | int(bytes[2])
	code = (code % 900000) + 100000

	return strconv.Itoa(code), nil
}

func (s *service) sendVerificationEmailAsync(email, code, username string, userID uint) {
	maxRetries := 3
	retryDelay := time.Second * 2

	for i := 0; i < maxRetries; i++ {
		if err := s.mailer.SendVerificationEmail(email, code, username); err != nil {
			s.log.Error("failed to send verification email",
				slog.String("error", err.Error()),
				slog.String("email", email),
				slog.Uint64("user_id", uint64(userID)),
				slog.Int("attempt", i+1))

			if i < maxRetries-1 {
				time.Sleep(retryDelay)
				retryDelay *= 2
			}
		} else {
			s.log.Info("verification email sent successfully",
				slog.String("email", email),
				slog.Uint64("user_id", uint64(userID)))
			return
		}
	}

	s.log.Error("failed to send verification email after all retries",
		slog.String("email", email),
		slog.Uint64("user_id", uint64(userID)))
}

func (s *service) Login(username, password string) (string, string, *model.User, error) {
	const op = "service.Login"
	log := s.log.With("op", op)

	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		log.Error("failed to get user by username", "error", err)
		return "", "", nil, errors.New(errors.CodeInvalidPassword, "Invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password))
	if err != nil {
		log.Error("invalid password", "error", err)
		return "", "", nil, errors.New(errors.CodeInvalidPassword, "Invalid username or password")
	}

	if !user.IsVerified {
		return "", "", nil, errors.New(errors.CodeUnauthorized, "Email not verified")
	}

	accessToken, err := jwtauth.GenerateAccessToken(user.ID, user.Username, s.cfg.JWTAccessSecret)
	if err != nil {
		log.Error("failed to generate access token", "error", err)
		return "", "", nil, err
	}

	refreshToken, err := jwtauth.GenerateRefreshToken(user.ID, user.RefreshTokenVersion, user.Username, s.cfg.JWTRefreshSecret)
	if err != nil {
		log.Error("failed to generate refresh token", "error", err)
		return "", "", nil, err
	}

	log.Debug("login successful", "user", user.ID)
	return accessToken, refreshToken, user, nil
}

func (s *service) LogoutFromAllSessions(userID uint) error {
	const op = "service.LogoutFromAllSessions"
	log := s.log.With("op", op)

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		log.Error("failed to get user by ID", "error", err)
		return err
	}

	user.RefreshTokenVersion += 1

	err = s.repo.SaveUser(user)
	if err != nil {
		log.Error("failed to save user", "error", err)
		return err
	}

	log.Debug("logout from all sessions", "user", user.ID)
	return nil
}

func (s *service) VerifyEmail(code string) (string, string, *model.User, error) {
	const op = "service.VerifyEmail"
	log := s.log.With("op", op)

	ctx := context.Background()
	userID, err := s.cache.GetUserIDByVerificationCode(ctx, code)
	if err != nil {
		log.Error("failed to get user ID by verification code", "error", err)
		return "", "", nil, err
	}

	if userID == 0 {
		return "", "", nil, errors.New(errors.CodeInvalidVerificationCode, "Invalid verification code")
	}

	isLimited, err := s.cache.IsVerificationRateLimited(ctx, userID)
	if err != nil {
		log.Error("failed to check verification rate limit", "error", err)
	} else if isLimited {
		log.Warn("verification rate limited", "user_id", userID)
		return "", "", nil, errors.New(errors.CodeTooManyRequests, "Too many verification attempts. Please try again later.")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		if trackErr := s.cache.TrackVerificationAttempt(ctx, userID); trackErr != nil {
			log.Error("failed to track verification attempt", "error", trackErr)
		}
		log.Error("failed to get user by ID", "error", err)
		return "", "", nil, err
	}

	if user.IsVerified {
		return "", "", nil, errors.New(errors.CodeInvalidRequest, "Email already verified")
	}

	user.IsVerified = true

	// Удаляем код верификации из Redis
	if err := s.cache.DeleteEmailVerificationCode(ctx, user.ID); err != nil {
		log.Error("failed to delete verification code", "error", err)
		// Не возвращаем ошибку, продолжаем процесс верификации
	}

	err = s.repo.SaveUser(user)
	if err != nil {
		log.Error("failed to save user after verification", "error", err)
		return "", "", nil, err
	}

	accessToken, err := jwtauth.GenerateAccessToken(user.ID, user.Username, s.cfg.JWTAccessSecret)
	if err != nil {
		log.Error("failed to generate access token", "error", err)
		return "", "", nil, err
	}

	refreshToken, err := jwtauth.GenerateRefreshToken(user.ID, user.RefreshTokenVersion, user.Username, s.cfg.JWTRefreshSecret)
	if err != nil {
		log.Error("failed to generate refresh token", "error", err)
		return "", "", nil, err
	}

	log.Debug("email verified", "user", user.ID)
	return accessToken, refreshToken, user, nil
}

func (s *service) SendEmailVerificationCode(email string) error {
	const op = "service.SendEmailVerificationCode"
	log := s.log.With("op", op)

	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		log.Error("failed to get user by email", "error", err)
		return err
	}

	if user.IsVerified {
		return errors.New(errors.CodeInvalidRequest, "Email already verified")
	}

	ctx := context.Background()

	// Проверяем существует ли уже код верификации
	code, err := s.cache.GetEmailVerificationCode(ctx, user.ID)
	if err != nil {
		log.Error("failed to check existing verification code", "error", err)
		return err
	}

	// Если кода нет, генерируем новый
	if code == "" {
		secureCode, err := s.generateSecureVerificationCode()
		if err != nil {
			log.Error("failed to generate secure verification code", "error", err)
			return errors.New(errors.CodeInternalError, "failed to generate verification code")
		}
		code = secureCode
		if err := s.cache.StoreEmailVerificationCode(ctx, user.ID, code); err != nil {
			log.Error("failed to store verification code", "error", err)
			return err
		}
	}

	go func() {
		if err := s.mailer.SendVerificationEmail(user.Email, code, user.Username); err != nil {
			log.Error("failed to send verification email", "error", err, "email", user.Email)
		}
	}()

	log.Debug("email verification code sent", "user", user.ID)
	return nil
}

func (s *service) RefreshTokens(refreshToken string) (string, string, error) {
	const op = "service.RefreshTokens"
	log := s.log.With("op", op)

	userID, err := jwtauth.ValidateRefreshToken(refreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		log.Error("failed to validate refresh token", "error", err)
		return "", "", errors.New(errors.CodeInvalidRefreshToken, "invalid refreshToken")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		if errors.IsErrorCode(err, errors.CodeInvalidRefreshToken) {
			return "", "", err
		}
		log.Error("failed to get user by refresh token", "error", err)
		return "", "", err
	}

	if user.RefreshTokenVersion != jwtauth.GetTokenVersion(refreshToken, s.cfg.JWTRefreshSecret) {
		return "", "", errors.New(errors.CodeInvalidRequest, "invalid refreshToken")
	}

	accessToken, err := jwtauth.GenerateAccessToken(user.ID, user.Username, s.cfg.JWTAccessSecret)
	if err != nil {
		log.Error("failed to generate access token", "error", err)
		return "", "", err
	}

	refreshToken, err = jwtauth.GenerateRefreshToken(user.ID, user.RefreshTokenVersion, user.Username, s.cfg.JWTRefreshSecret)
	if err != nil {
		log.Error("failed to generate refresh token", "error", err)
		return "", "", err
	}

	err = s.repo.SaveUser(user)
	if err != nil {
		log.Error("failed to save user", "error", err)
		return "", "", err
	}

	log.Debug("tokens refreshed", "user", user.ID)
	return accessToken, refreshToken, nil
}

func (s *service) RequestPasswordReset(email string) error {
	const op = "service.RequestPasswordReset"
	log := s.log.With("op", op)

	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		log.Error("failed to get user by email", "error", err)
		return err
	}

	token, err := generateResetToken()
	if err != nil {
		log.Error("failed to generate reset token", "error", err)
		return err
	}

	// Сохраняем токен в репозитории
	ctx := context.Background()
	err = s.cache.StoreResetToken(ctx, token, user.ID)
	if err != nil {
		log.Error("failed to store reset token", "error", err)
		return errors.New(errors.CodeInternalError, "Failed to process password reset")
	}

	// Отправляем письмо со ссылкой для сброса пароля
	go func() {
		if err := s.mailer.SendResetEmail(user.Email, user.Username, token); err != nil {
			log.Error("failed to send reset email", "error", err)
		}
	}()

	return nil
}

func (s *service) ResetPassword(token, password string) error {
	const op = "service.ResetPassword"
	log := s.log.With("op", op)

	ctx := context.Background()
	userID, err := s.cache.GetUserIDByResetToken(ctx, token)
	if err != nil {
		log.Error("failed to get user ID by reset token", "error", err)
		return errors.New(errors.CodeInternalError, "Failed to process password reset")
	}

	if userID == 0 {
		return errors.New(errors.CodeInvalidRequest, "Invalid or expired reset token")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		log.Error("failed to get user by ID", "error", err)
		return err
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", "error", err)
		return errors.New(errors.CodeInternalError, "Failed to reset password")
	}

	user.PassHash = string(hashPassword)
	user.RefreshTokenVersion++

	err = s.repo.SaveUser(user)
	if err != nil {
		log.Error("failed to save user", "error", err)
		return err
	}

	err = s.cache.DeleteResetToken(ctx, token)
	if err != nil {
		log.Error("failed to delete reset token", "error", err)
	}

	return nil
}

// generateResetToken генерирует уникальный токен для сброса пароля
func generateResetToken() (string, error) {
	b := make([]byte, 32)
	_, err := cryptorand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure random token: %w", err)
	}
	return fmt.Sprintf("%x", b), nil
}

func (s *service) AddBookmark(userID uint, title, url string, showText bool) (*model.Bookmark, error) {
	const op = "service.AddBookmark"
	log := s.log.With("op", op)

	ctx := context.Background()
	faviconBase64, err := parsers.FetchFaviconBase64(ctx, s.cache, url)
	if err != nil {
		log.Error("failed to fetch favicon", "error", err, "url", url)
	}

	bookmark := &model.Bookmark{
		UserID:    userID,
		Title:     title,
		URL:       url,
		ShowText:  showText,
		Favicon:   faviconBase64,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.AddBookmark(bookmark)
	if err != nil {
		log.Error("failed to add bookmark", "error", err, "user_id", userID)
		return nil, err
	}

	log.Debug("bookmark added successfully", "bookmark_id", bookmark.ID, "user_id", userID)
	return bookmark, nil
}

func (s *service) GetBookmarks(userID uint) ([]model.Bookmark, error) {
	const op = "service.GetBookmarks"
	log := s.log.With("op", op)

	bookmarks, err := s.repo.GetBookmarks(userID)
	if err != nil {
		log.Error("failed to get bookmarks", "error", err, "user_id", userID)
		return nil, err
	}

	log.Debug("bookmarks retrieved successfully", "user_id", userID, "count", len(bookmarks))
	return bookmarks, nil
}

func (s *service) GetBookmarkByID(userID, bookmarkID uint) (*model.Bookmark, error) {
	const op = "service.GetBookmarkByID"
	log := s.log.With("op", op)

	bookmark, err := s.repo.GetBookmarkByID(bookmarkID)
	if err != nil {
		log.Error("failed to get bookmark by ID", "error", err, "bookmark_id", bookmarkID)
		return nil, err
	}

	if bookmark.UserID != userID {
		log.Error("bookmark doesn't belong to user", "user_id", userID, "bookmark_id", bookmarkID, "bookmark_user_id", bookmark.UserID)
		return nil, errors.New(errors.CodeForbidden, "Bookmark doesn't belong to user")
	}

	log.Debug("bookmark retrieved successfully", "bookmark_id", bookmarkID, "user_id", userID)
	return bookmark, nil
}

func (s *service) PatchBookmark(userID, bookmarkID uint, patch *model.PatchBookmarkRequest) (*model.Bookmark, error) {
	const op = "service.PatchBookmark"
	log := s.log.With("op", op)

	bookmark, err := s.GetBookmarkByID(userID, bookmarkID)
	if err != nil {
		log.Error("failed to get bookmark for update", "error", err, "bookmark_id", bookmarkID, "user_id", userID)
		return nil, err
	}

	if patch.Title != nil {
		bookmark.Title = *patch.Title
	}
	if patch.URL != nil {
		bookmark.URL = *patch.URL
		ctx := context.Background()
		faviconBase64, err := parsers.FetchFaviconBase64(ctx, s.cache, *patch.URL)
		if err != nil {
			log.Error("failed to fetch favicon", "error", err, "url", *patch.URL)
		}
		bookmark.Favicon = faviconBase64
	}
	if patch.ShowText != nil {
		bookmark.ShowText = *patch.ShowText
	}
	bookmark.UpdatedAt = time.Now()

	err = s.repo.UpdateBookmark(bookmark)
	if err != nil {
		log.Error("failed to update bookmark", "error", err, "bookmark_id", bookmarkID)
		return nil, err
	}

	log.Debug("bookmark updated successfully", "bookmark_id", bookmarkID, "user_id", userID)
	return bookmark, nil
}

func (s *service) DeleteBookmark(userID, bookmarkID uint) error {
	const op = "service.DeleteBookmark"
	log := s.log.With("op", op)

	bookmark, err := s.GetBookmarkByID(userID, bookmarkID)
	if err != nil {
		log.Error("failed to get bookmark for deletion", "error", err, "bookmark_id", bookmarkID, "user_id", userID)
		return err
	}

	err = s.repo.DeleteBookmark(bookmark.ID)
	if err != nil {
		log.Error("failed to delete bookmark", "error", err, "bookmark_id", bookmarkID)
		return err
	}

	log.Debug("bookmark deleted successfully", "bookmark_id", bookmarkID, "user_id", userID)
	return nil
}

func (s *service) ImportBookmarks(userID uint, base64Data string) ([]model.Bookmark, error) {
	const op = "service.ImportBookmarks"
	log := s.log.With("op", op)

	ctx := context.Background()

	parsedBookmarks, err := parsers.ParseBookmarksFromHTML(ctx, base64Data, s.cache)
	if err != nil {
		log.Error("failed to parse bookmarks HTML", "error", err, "user_id", userID)
		return nil, errors.New(errors.CodeInvalidRequest, "Failed to parse bookmarks file")
	}

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		log.Error("failed to get user", "error", err, "user_id", userID)
		return nil, errors.New(errors.CodeInvalidRequest, "Failed to get user")
	}

	if int(user.AmountOfBookmarks)-len(parsedBookmarks) < 0 {
		log.Error("reached maximum bookmarks", "user_id", userID, "amount", user.AmountOfBookmarks)
		return nil, errors.New(errors.CodeInvalidRequest, "Reached maximum bookmarks")
	}

	now := time.Now()
	savedBookmarks := make([]model.Bookmark, 0, len(parsedBookmarks))

	for _, bookmark := range parsedBookmarks {
		bookmark.UserID = userID
		bookmark.CreatedAt = now
		bookmark.UpdatedAt = now

		err = s.repo.AddBookmark(&bookmark)
		if err != nil {
			log.Error("failed to save imported bookmark", "error", err, "user_id", userID, "url", bookmark.URL)
			continue
		}

		savedBookmarks = append(savedBookmarks, bookmark)
	}

	log.Debug("bookmarks imported successfully", "user_id", userID, "count", len(savedBookmarks))
	return savedBookmarks, nil
}

func (s *service) ExportBookmarks(userID uint) (string, error) {
	const op = "service.ExportBookmarks"
	log := s.log.With("op", op)

	bookmarks, err := s.repo.GetBookmarks(userID)
	if err != nil {
		log.Error("failed to get bookmarks for export", "error", err, "user_id", userID)
		return "", err
	}

	htmlBase64, err := parsers.ExportBookmarksToHTML(bookmarks)
	if err != nil {
		log.Error("failed to export bookmarks to HTML", "error", err, "user_id", userID)
		return "", errors.New(errors.CodeInternalError, "Failed to export bookmarks")
	}

	log.Debug("bookmarks exported successfully", "user_id", userID, "count", len(bookmarks))
	return htmlBase64, nil
}

func (s *service) ImportBookmarksV2(userID uint, bookmarks []model.BookmarkV2Request) ([]model.Bookmark, error) {
	const op = "service.ImportBookmarksV2"
	log := s.log.With("op", op)

	importedBookmarks := make([]model.Bookmark, 0, len(bookmarks))

	for i, bookmark := range bookmarks {
		importedBookmarks = append(importedBookmarks, model.Bookmark{
			UserID:    userID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Title:     bookmark.Title,
			URL:       bookmark.URL,
			ShowText:  bookmark.ShowText,
			Favicon:   bookmark.Favicon,
		})

		if bookmark.Favicon == "" {
			ctx := context.Background()
			faviconBase64, err := parsers.FetchFaviconBase64(ctx, s.cache, bookmark.URL)
			if err != nil {
				log.Error("failed to fetch favicon", "error", err, "url", bookmark.URL)
			}
			importedBookmarks[i].Favicon = faviconBase64
		}

		err := s.repo.AddBookmark(&importedBookmarks[i])
		if err != nil {
			log.Error("failed to add bookmark", "error", err, "user_id", userID, "url", bookmark.URL)
			return nil, err
		}
	}

	return importedBookmarks, nil
}

func (s *service) ExportBookmarksV2(userID uint) ([]model.Bookmark, error) {
	const op = "service.ExportBookmarksV2"
	log := s.log.With("op", op)

	bookmarks, err := s.repo.GetBookmarks(userID)
	if err != nil {
		log.Error("failed to get bookmarks for export", "error", err, "user_id", userID)
		return nil, err
	}

	return bookmarks, nil
}

func (s *service) GetUser(userID any) (*model.UserResponse, error) {
	const op = "service.GetUser"
	log := s.log.With("op", op)

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		log.Error("failed to get user", "error", err, "user_id", userID)
		return nil, err
	}

	userResp := model.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		IsPremium: user.IsPremium,
	}

	return &userResp, nil
}
