package handlers

import (
	"log/slog"

	"github.com/aerscs/theca-public/internal/model"
	"github.com/aerscs/theca-public/internal/service"
	errors "github.com/aerscs/theca-public/internal/utils/errors"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
	log     *slog.Logger
}

func NewHandler(service service.Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// @Summary Health Check
// @Description Check if the service is healthy
// @Tags health
// @Produce json
// @Success 200
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})

	h.log.Debug("health check successful")
}

// @Summary Register
// @Description Register a new user
// @Tags user
// @Accept json
// @Produce json
// @Param registerRequest body model.RegisterRequest true "Register request"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /v1/register [post]
func (h *Handler) Register(c *gin.Context) {
	const op = "handler.register"
	log := h.log.With(slog.String("op", op))

	var req model.RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		log.Debug("invalid request format", slog.String("error", err.Error()))
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "invalid request format"))
		return
	}

	userID, err := h.service.Register(&req)
	if err != nil {
		log.Error("failed to register user", slog.String("error", err.Error()), slog.String("username", req.Username))
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("user registration successful", slog.Uint64("user_id", uint64(userID)), slog.String("username", req.Username))
	errors.RespondWithSuccess(c, model.RegisterResponse{
		Message: "User registered successfully",
	})
}

// @Summary Login
// @Description Login a user
// @Tags user
// @Accept json
// @Produce json
// @Param loginRequest body model.LoginRequest true "Login request"
// @Success 200 {object} model.LoginResponse
// @Failure 400
// @Failure 500
// @Router /v1/login [post]
func (h *Handler) Login(c *gin.Context) {
	const op = "handler.login"
	log := h.log.With(slog.String("op", op))

	var req model.LoginRequest
	if err := c.BindJSON(&req); err != nil {
		log.Debug("binding json", "err", err, "req", req)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "invalid request format"))
		return
	}

	accessToken, refreshToken, user, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		log.Error("failed to login user", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	c.SetCookie("refreshToken", refreshToken, 0, "/", "", false, true)

	log.Debug("user login handled successfully", "username", req.Username)
	errors.RespondWithSuccess(c, model.LoginResponse{
		AccessToken: accessToken,
		User: model.UserResponse{
			Username:  user.Username,
			Email:     user.Email,
			ID:        user.ID,
			IsPremium: user.IsPremium,
		},
	})
}

// @Summary Logout
// @Description Logout a user
// @Tags user
// @Accept json
// @Produce json
// @Success 200
// @Failure 400
// @Failure 500
// @Failure 401
// @Security Bearer
// @Router /v1/api/logout [delete]
func (h *Handler) Logout(c *gin.Context) {
	const op = "handler.logout"
	log := h.log.With(slog.String("op", op))

	c.SetCookie("refreshToken", "", -1, "/", "", false, true)
	log.Debug("user logout handled successfully", "user", c.GetUint("user_id"))
	errors.RespondWithSuccess(c, "Logged out successfully")
}

// @Summary Verify Email
// @Description Verify email
// @Tags user
// @Accept json
// @Produce json
// @Param emailVerifyRequest body model.EmailVerifyRequest true "Email verify request"
// @Success 200 {object} model.LoginResponse
// @Failure 400
// @Failure 500
// @Router /v1/verify-email [patch]
func (h *Handler) VerifyEmail(c *gin.Context) {
	const op = "handler.verifyEmail"
	log := h.log.With(slog.String("op", op))

	var req model.EmailVerifyRequest
	if err := c.BindJSON(&req); err != nil {
		log.Debug("binding json", "err", err, "req", req)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "invalid request format"))
		return
	}

	accessToken, refreshToken, user, err := h.service.VerifyEmail(req.Code)
	if err != nil {
		log.Error("failed to verify email", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	c.SetCookie("refreshToken", refreshToken, 0, "/", "", false, true)

	log.Debug("email verification handled successfully", "code", req.Code)
	errors.RespondWithSuccess(c, model.LoginResponse{
		AccessToken: accessToken,
		User: model.UserResponse{
			Username:  user.Username,
			Email:     user.Email,
			ID:        user.ID,
			IsPremium: user.IsPremium,
		},
	})
}

// @Summary Send Email Verification Code
// @Description Send email verification code
// @Tags user
// @Accept json
// @Produce json
// @Param sendEmailCodeRequest body model.SendEmailVerificationCodeRequest true "SendEmailVerificationCodeRequest request"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /v1/send-email-verification-code [post]
func (h *Handler) SendEmailVerificationCode(c *gin.Context) {
	const op = "handler.sendEmailVerificationCode"
	log := h.log.With(slog.String("op", op))

	var req model.SendEmailVerificationCodeRequest
	if err := c.BindJSON(&req); err != nil {
		log.Info("failed to parse request", "error", err)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "invalid userID"))
		return
	}

	err := h.service.SendEmailVerificationCode(req.Email)
	if err != nil {
		log.Error("failed to send email verification code", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("email verification code sending handled successfully", "user", req.Email)
	errors.RespondWithSuccess(c, "Email verification code sent successfully")
}

// @Summary Refresh Tokens
// @Description Refresh tokens
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} model.LoginResponse
// @Failure 400
// @Failure 401
// @Failure 500
// @Router /v1/refresh-tokens [get]
func (h *Handler) RefreshTokens(c *gin.Context) {
	const op = "handler.RefreshTokens"
	log := h.log.With(slog.String("op", op))

	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		log.Error("failed to get refresh token from cookie", "error", err)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "invalid refreshToken"))
		return
	}

	accessToken, refreshToken, err := h.service.RefreshTokens(refreshToken)
	if err != nil {
		log.Error("failed to refresh tokens", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	c.SetCookie("refreshToken", refreshToken, 0, "/", "", false, true)

	log.Debug("token refresh handled successfully")
	errors.RespondWithSuccess(c, gin.H{
		"access_token": accessToken,
	})
}

// @Summary Request Password Reset
// @Description Send email with password reset link
// @Tags user
// @Accept json
// @Produce json
// @Param passwordResetRequest body model.PasswordResetRequest true "Password reset request"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /v1/request-password-reset [post]
func (h *Handler) RequestPasswordReset(c *gin.Context) {
	const op = "handler.requestPasswordReset"
	log := h.log.With(slog.String("op", op))

	var req model.PasswordResetRequest
	if err := c.BindJSON(&req); err != nil {
		log.Debug("binding json", "err", err, "req", req)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "invalid request format"))
		return
	}

	err := h.service.RequestPasswordReset(req.Email)
	if err != nil {
		log.Error("failed to request password reset", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("password reset requested successfully", "email", req.Email)
	errors.RespondWithSuccess(c, "Password reset link has been sent to your email")
}

// @Summary Reset Password
// @Description Reset password using token from email
// @Tags user
// @Accept json
// @Produce json
// @Param resetToken query string true "Reset token"
// @Param resetPasswordRequest body model.ResetPasswordRequest true "Reset password request"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /v1/reset-password [patch]
func (h *Handler) ResetPassword(c *gin.Context) {
	const op = "handler.resetPassword"
	log := h.log.With(slog.String("op", op))

	token := c.Query("token")
	if token == "" {
		log.Debug("missing reset token")
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "reset token is required"))
		return
	}

	var req model.ResetPasswordRequest
	if err := c.BindJSON(&req); err != nil {
		log.Debug("binding json", "err", err, "req", req)
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "invalid request format"))
		return
	}

	err := h.service.ResetPassword(token, req.Password)
	if err != nil {
		log.Error("failed to reset password", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("password reset successfully")
	errors.RespondWithSuccess(c, "Password has been reset successfully")
}

// @Summary Get yourself
// @Description Get user information
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} model.UserResponse
// @Failure 400
// @Failure 500
// @Router /v1/api/user/me [get]
func (h *Handler) GetSelfUser(c *gin.Context) {
	const op = "handler.getUser"
	log := h.log.With(slog.String("op", op))

	userID, ok := c.Get("userID")
	if !ok {
		log.Debug("missing user id")
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "user id is required"))
		return
	}

	user, err := h.service.GetUser(userID)
	if err != nil {
		log.Error("failed to get user", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("user retrieved successfully", "user", user)
	errors.RespondWithSuccess(c, user)
}

// @Summary Get user by ID
// @Description Get user information
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} model.UserResponse
// @Failure 400
// @Failure 500
// @Router /v1/api/user/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	const op = "handler.getUser"
	log := h.log.With(slog.String("op", op))

	userID, ok := c.Params.Get("id")
	if !ok {
		log.Debug("missing user id")
		errors.RespondWithError(c, errors.New(errors.CodeInvalidRequest, "user id is required"))
		return
	}

	user, err := h.service.GetUser(userID)
	if err != nil {
		log.Error("failed to get user", "error", err)
		errors.RespondWithError(c, err)
		return
	}

	log.Debug("user retrieved successfully", "user", user)
	errors.RespondWithSuccess(c, user)
}
