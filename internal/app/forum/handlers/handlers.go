package handlers

import (
	customErr "DBForum/internal/app/errors"
	forumUseCase "DBForum/internal/app/forum/usecase"
	"DBForum/internal/app/httputils"
	"DBForum/internal/app/models"
	"errors"
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
)

type Handlers struct {
	useCase forumUseCase.UseCase
}

func NewHandler(useCase forumUseCase.UseCase) *Handlers {
	return &Handlers{
		useCase: useCase,
	}
}

func (h *Handlers) Create(ctx *fasthttp.RequestCtx) {
	forum := &models.Forum{}

	if err := easyjson.Unmarshal(ctx.PostBody(), forum); err != nil {
		log.Println(err)
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		return
	}

	var err error
	nickname := forum.User
	forum, err = h.useCase.CreateForum(forum)
	if errors.Is(err, customErr.ErrUserNotFound) {
		resp := map[string]string{
			"message": "Can't find user with nickname: " + nickname,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrDuplicate) {
		httputils.Respond(ctx, http.StatusConflict, forum)
		return
	}
	if err != nil {
		log.Println(err)
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		return
	}
	httputils.Respond(ctx, http.StatusCreated, forum)
}

func (h *Handlers) Details(ctx *fasthttp.RequestCtx) {
	slug := ctx.UserValue("slug").(string)
	forum, err := h.useCase.GetInfoBySlug(slug)
	if errors.Is(err, customErr.ErrForumNotFound) {
		resp := map[string]string{
			"message": "Can't find forum with slug: " + slug,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(ctx, http.StatusOK, forum)
}

func (h *Handlers) CreateThread(ctx *fasthttp.RequestCtx) {
	thread := &models.Thread{}
	if err := easyjson.Unmarshal(ctx.PostBody(), thread); err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	forumSlug := ctx.UserValue("slug").(string)
	nickname := thread.Author
	thread.Forum = forumSlug

	var err error
	thread, err = h.useCase.CreateThread(thread)
	if errors.Is(err, customErr.ErrUserNotFound) {
		resp := map[string]string{
			"message": "Can't find thread author by nickname: " + nickname,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrForumNotFound) {
		resp := map[string]string{
			"message": "Can't find thread forum by slug: " + forumSlug,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrDuplicate) {
		httputils.Respond(ctx, http.StatusConflict, thread)
		return
	}
	if err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(ctx, http.StatusCreated, thread)
}

func (h *Handlers) GetUsers(ctx *fasthttp.RequestCtx) {
	forumSlug := ctx.UserValue("slug").(string)
	// максимальное количество возвращаемых записей
	limit := ctx.QueryArgs().GetUintOrZero("limit")
	// Идентификатор пользователя, с которого будут выводиться пользоватли
	//(пользователь с данным идентификатором в результат не попадает).

	since := string(ctx.QueryArgs().Peek("since"))
	// Флаг сортировки по убыванию.
	desc := ctx.QueryArgs().GetBool("desc")

	var users models.UserList
	var err error
	users, err = h.useCase.GetForumUsers(forumSlug, limit, since, desc)
	if errors.Is(err, customErr.ErrForumNotFound) {
		resp := map[string]string{
			"message": "Can't find forum by slug: " + forumSlug,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	httputils.Respond(ctx, http.StatusOK, users)
}

func (h *Handlers) GetThreads(ctx *fasthttp.RequestCtx) {
	forumSlug := ctx.UserValue("slug").(string)
	var threads models.ThreadList

	// максимальное количество возвращаемых записей
	limit, _ := ctx.QueryArgs().GetUint("limit")
	// Дата создания ветви обсуждения, с которой будут выводиться записи
	// (ветвь обсуждения с указанной датой попадает в результат выборки).
	since := string(ctx.QueryArgs().Peek("since"))
	// Флаг сортировки по убыванию.
	desc := ctx.QueryArgs().GetBool("desc")

	var err error
	threads, err = h.useCase.GetForumThreads(forumSlug, limit, since, desc)
	if errors.Is(err, customErr.ErrForumNotFound) {
		resp := map[string]string{
			"message": "Can't find forum by slug: " + forumSlug,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(ctx, http.StatusOK, threads)
}
