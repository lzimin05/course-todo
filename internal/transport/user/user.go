package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/lzimin05/course-todo/config"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/user"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/transport/utils/response"
	validation "github.com/lzimin05/course-todo/internal/transport/utils/validation/user"
)

//go:generate mockgen -source=user.go -destination=../../usecase/mocks/user_usecase_mock.go -package=mocks IUserUsecase
type IUserUsecase interface {
	GetMe(context.Context) (*dto.UserDTO, error)
	GetUserByEmail(context.Context, string) (*dto.UserDTO, error)
	GetUserByLogin(context.Context, string) (*dto.UserDTO, error)
	UpdateUsername(context.Context, string) error
}

type UserHandler struct {
	uc     IUserUsecase
	config *config.Config
}

func New(uc IUserUsecase, conf *config.Config) *UserHandler {
	return &UserHandler{
		uc:     uc,
		config: conf,
	}
}

// GetMe возвращает информацию о текущем пользователе
// @Summary      Получить информацию о текущем пользователе
// @Description  Возвращает информацию о текущем авторизованном пользователе
// @Tags         user
// @Produce      json
// @Success      200  {object} dto.UserDTO "Информация о пользователе"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Security     BearerAuth
// @Router       /users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetMe"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	user, err := h.uc.GetMe(r.Context())
	if err != nil {
		logger.WithError(err).Warn("err")
		response.SendError(r.Context(), w, http.StatusBadRequest, err.Error())
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, user)
}

// GetUserByEmail возвращает информацию о пользователе по email
// @Summary      Получить пользователя по email
// @Description  Возвращает информацию о пользователе по его email адресу
// @Tags         user
// @Param        email   query     string  true  "Email пользователя"
// @Produce      json
// @Success      200  {object} dto.UserDTO "Информация о пользователе"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      404  {object} dto.ErrorResponse "Пользователь не найден"
// @Security     BearerAuth
// @Router       /users/by-email [get]
func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetUserByEmail"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	email := r.URL.Query().Get("email")
	if email == "" {
		logger.Warn("email parameter is required")
		response.SendError(r.Context(), w, http.StatusBadRequest, "email parameter is required")
		return
	}

	user, err := h.uc.GetUserByEmail(r.Context(), email)
	if err != nil {
		logger.WithError(err).Warn("err")
		response.SendError(r.Context(), w, http.StatusNotFound, "user not found")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, user)
}

// GetUserByLogin возвращает информацию о пользователе по login
// @Summary      Получить пользователя по логину
// @Description  Возвращает информацию о пользователе по его логину
// @Tags         user
// @Param        login   query     string  true  "Логин пользователя"
// @Produce      json
// @Success      200  {object} dto.UserDTO "Информация о пользователе"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      404  {object} dto.ErrorResponse "Пользователь не найден"
// @Security     BearerAuth
// @Router       /users/by-login [get]
func (h *UserHandler) GetUserByLogin(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.GetUserByLogin"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	login := r.URL.Query().Get("login")
	if login == "" {
		logger.Warn("login parameter is required")
		response.SendError(r.Context(), w, http.StatusBadRequest, "login parameter is required")
		return
	}

	user, err := h.uc.GetUserByLogin(r.Context(), login)
	if err != nil {
		logger.WithError(err).Warn("err")
		response.SendError(r.Context(), w, http.StatusNotFound, "user not found")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, user)
}

// UpdateUsername обновляет имя пользователя
// @Summary      Обновить имя пользователя
// @Description  Обновляет имя пользователя для текущего авторизованного пользователя
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        request body dto.UpdateUsernameRequest true "Новое имя пользователя"
// @Success      200  "Имя пользователя обновлено"
// @Failure      400  {object} dto.ErrorResponse "Неверный запрос"
// @Failure      401  {object} dto.ErrorResponse "Пользователь не авторизован"
// @Security     BearerAuth
// @Router       /users/username [patch]
func (h *UserHandler) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	const op = "UserHandler.UpdateUsername"
	logger := logctx.GetLogger(r.Context()).WithField("op", op)

	var req dto.UpdateUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Warn("failed to decode request")
		response.SendError(r.Context(), w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Валидация запроса
	if err := validation.ValidateUpdateUsernameRequest(req); err != nil {
		logger.WithError(err).Warn("validation failed")
		response.SendError(r.Context(), w, http.StatusBadRequest, err.Error())
		return
	}

	err := h.uc.UpdateUsername(r.Context(), req.Username)
	if err != nil {
		logger.WithError(err).Error("failed to update username")
		response.SendError(r.Context(), w, http.StatusInternalServerError, "failed to update username")
		return
	}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, nil)
}
