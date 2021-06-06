package handlers

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/httputils"
	"DBForum/internal/app/models"
	userUseCase "DBForum/internal/app/user/usecase"
	"errors"
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
)

type Handlers struct {
	useCase userUseCase.UseCase
}

func NewHandler(useCase userUseCase.UseCase) *Handlers {
	return &Handlers{
		useCase: useCase,
	}
}

func (h *Handlers) CreateUser(ctx *fasthttp.RequestCtx) {
	nickname := ctx.UserValue("nickname").(string)

	user := models.User{Nickname: nickname}
	if err := easyjson.Unmarshal(ctx.PostBody(), &user); err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	err := h.useCase.CreateUser(user)
	if errors.Is(err, customErr.ErrDuplicate) {
		var users models.UserList
		users, err = h.useCase.GetUsersByNickAndEmail(user.Nickname, user.Email)
		if err != nil {
			httputils.Respond(ctx, http.StatusInternalServerError, nil)
			log.Println(err)
			return
		}
		httputils.Respond(ctx, http.StatusConflict, users)
		return
	}
	if err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(ctx, http.StatusCreated, user)
}

func (h *Handlers) GetUserInfo(ctx *fasthttp.RequestCtx) {
	nickname := ctx.UserValue("nickname").(string)
	user := &models.User{Nickname: nickname}

	user, err := h.useCase.GetUserInfo(nickname)

	if errors.Is(err, customErr.ErrUserNotFound) {
		resp := map[string]string{
			"message": "Can't find user by nickname: " + nickname,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	httputils.Respond(ctx, http.StatusOK, user)
}

func (h *Handlers) ChangeUser(ctx *fasthttp.RequestCtx) {
	nickname := ctx.UserValue("nickname").(string)

	user := models.User{Nickname: nickname}
	if err := easyjson.Unmarshal(ctx.PostBody(), &user); err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	err := h.useCase.ChangeUser(&user)
	if errors.Is(err, customErr.ErrUserNotFound) {
		resp := map[string]string{
			"message": "Can't find user by nickname: " + nickname,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrConflict) {
		resp := map[string]string{
			"message": "This email is already registered by user: ",
		}
		httputils.RespondErr(ctx, http.StatusConflict, resp)
		return
	}
	httputils.Respond(ctx, http.StatusOK, user)
}
