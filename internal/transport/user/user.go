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
