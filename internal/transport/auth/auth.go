package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/lzimin05/course-todo/config"
	"github.com/lzimin05/course-todo/internal/models/domains"
	errs "github.com/lzimin05/course-todo/internal/models/errs"
	"github.com/lzimin05/course-todo/internal/transport/dto"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/transport/utils/cookie"
	response "github.com/lzimin05/course-todo/internal/transport/utils/response"
)

//go:generate mockgen -source=auth.go -destination=../../usecase/mocks/auth_usecase_mock.go -package=mocks AuthUsecase
type AuthUsecase interface {
	Authenticate(ctx context.Context, login_or_email, password string) (string, error)
	Register(ctx context.Context, login, username, email, password string) (string, error)
	Logout(ctx context.Context, token string) error
}

type AuthHandler struct {
	uc     AuthUsecase
	config *config.Config
}

func New(uc AuthUsecase, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		uc:     uc,
		config: cfg,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "AuthHandler.Login"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode login request")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	token, err := h.uc.Authenticate(r.Context(), req.EmailOrLogin, req.Password)
	if err != nil {
		logger.WithError(err).Warn("authentication failed")
		response.SendError(r.Context(), w, http.StatusUnauthorized, "Incorrect data")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})

	respTok := dto.TokenResponse{Token: token}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, respTok)
}

// Register регистрирует нового пользователя
// @Summary      Регистрация пользователя
// @Description  Регистрирует нового пользователя в системе и создает сессию
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        user  body  dto.RegisterRequest  true  "Данные для регистрации"
// @Success      201  "Пользователь успешно зарегистрирован"
// @Failure      400  "Ошибки регистрации"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "AuthHandler.Register"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode registration request")
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	token, err := h.uc.Register(r.Context(), req.Login, req.Username, req.Email, req.Password)
	if err != nil {
		if err == errs.ErrIsDuplicateKey {
			logger.WithError(err).Warn("user with this login or email already exists")
			response.SendError(r.Context(), w, http.StatusConflict, "user with this login or email already exists")
			return
		}
		logger.WithError(err).Warn("registration failed")
		response.SendError(r.Context(), w, http.StatusBadRequest, "error registation")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})

	response.SendJSONResponse(r.Context(), w, http.StatusCreated, nil)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	const op = "AuthHandler.Logout"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	jwtCookie, err := r.Cookie(string(domains.TokenCookieName))
	if err != nil {
		logger.WithError(err).Warn("JWT token required")
		response.SendError(r.Context(), w, http.StatusUnauthorized, "JWT token required")
		return
	}

	err = h.uc.Logout(r.Context(), jwtCookie.Value)
	if err != nil {
		logger.WithError(err).Warn("error logout")
		response.SendError(r.Context(), w, http.StatusInternalServerError, err.Error())
		return
	}

	cookieProvider := cookie.NewCookieProvider(h.config)
	cookieProvider.Unset(w, domains.TokenCookieName)

	response.SendJSONResponse(r.Context(), w, http.StatusOK, nil)
}
