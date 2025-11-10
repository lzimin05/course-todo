package transport

import (
	"context"
	"net/http"

	"github.com/lzimin05/course-todo/config"
	dto "github.com/lzimin05/course-todo/internal/transport/dto/user"
	"github.com/lzimin05/course-todo/internal/transport/middleware/logctx"
	"github.com/lzimin05/course-todo/internal/transport/utils/response"
)

type IUserUsecase interface {
	GetMe(context.Context) (*dto.UserDTO, error)
	GetUserByEmail(context.Context, string) (*dto.UserDTO, error)
	GetUserByLogin(context.Context, string) (*dto.UserDTO, error)
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
