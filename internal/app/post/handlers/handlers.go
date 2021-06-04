package handlers

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/httputils"
	"DBForum/internal/app/models"
	postUseCase "DBForum/internal/app/post/usecase"
	"errors"
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Handlers struct {
	useCase postUseCase.UseCase
}

func NewHandler(useCase postUseCase.UseCase) *Handlers {
	return &Handlers{
		useCase: useCase,
	}
}

func (h *Handlers) GetInfo(ctx *fasthttp.RequestCtx) {
	id, _ := strconv.ParseUint(ctx.UserValue("id").(string), 10, 64)
	// Включение полной информации о соответвующем объекте сообщения.
	// Если тип объекта не указан, то полная информация об этих объектах не
	// передаётся.
	// values: user/forum/thread
	related := strings.Split(string(ctx.QueryArgs().Peek("related")), ",")

	postInfo, err := h.useCase.GetPostInfoByID(id, related)

	if errors.Is(err, customErr.ErrPostNotFound) {
		resp := map[string]string{
			"message": "Can't find post with id: ",
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		log.Println(err)
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		return
	}
	httputils.Respond(ctx, http.StatusOK, postInfo)
}

func (h *Handlers) ChangeMessage(ctx *fasthttp.RequestCtx) {
	post := &models.Post{}
	if err := easyjson.Unmarshal(ctx.PostBody(), post); err != nil {
		log.Println(err)
		httputils.Respond(ctx, http.StatusInternalServerError, post)
		return
	}
	id, _ := strconv.ParseUint(ctx.UserValue("id").(string), 10, 64)

	post.ID = id
	var err error
	post, err = h.useCase.ChangeMessage(*post)
	if errors.Is(err, customErr.ErrPostNotFound) {
		resp := map[string]string{
			"message": "Can't find post with id: " + strconv.FormatUint(id, 10),
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		log.Println(err)
		httputils.Respond(ctx, http.StatusInternalServerError, post)
		return
	}
	httputils.Respond(ctx, http.StatusOK, post)
}
