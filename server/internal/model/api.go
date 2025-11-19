package model

import "time"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type EmailVerifyRequest struct {
	Code string `json:"code" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	ID        uint   `json:"id"`
	IsPremium bool   `json:"is_premium"`
}

type ChangePasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

type RequestPasswordReset struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPassword struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RequestVerificationToken struct {
	Username string `json:"username" binding:"required,min=3"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

// AddBookmarkRequest запрос на добавление закладки
type AddBookmarkRequest struct {
	Title    string `json:"title" binding:"required"`
	URL      string `json:"url" binding:"required"`
	ShowText bool   `json:"show_text"`
}

// UpdateBookmarkRequest запрос на обновление закладки
type UpdateBookmarkRequest struct {
	Title    string `json:"title" binding:"required"`
	URL      string `json:"url" binding:"required"`
	ShowText bool   `json:"show_text" binding:"required"`
}

// PatchBookmarkRequest запрос на частичное обновление закладки
type PatchBookmarkRequest struct {
	Title    *string `json:"title,omitempty"`
	URL      *string `json:"url,omitempty"`
	ShowText *bool   `json:"show_text,omitempty"`
}

// BookmarkResponse ответ с данными закладки
type BookmarkResponse struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Favicon   string    `json:"favicon"`
	ID        uint      `json:"id"`
	ShowText  bool      `json:"show_text"`
}

type SendEmailVerificationCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}
