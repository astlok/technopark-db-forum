package handlers

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/httputils"
	"DBForum/internal/app/models"
	threadUseCase "DBForum/internal/app/thread/usecase"
	"github.com/mailru/easyjson"
	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"strconv"
)

type Handlers struct {
	useCase threadUseCase.UseCase
}

func NewHandler(useCase threadUseCase.UseCase) *Handlers {
	return &Handlers{
		useCase: useCase,
	}
}

func (h *Handlers) CreatePost(ctx *fasthttp.RequestCtx) {
	var posts models.PostList
	if err := easyjson.Unmarshal(ctx.PostBody(), &posts); err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	idOrSlug := ctx.UserValue("slug_or_id").(string)
	posts, err := h.useCase.CreatePosts(idOrSlug, posts)
	if errors.Is(err, customErr.ErrThreadNotFound) {
		var message string
		if _, err := strconv.ParseUint(idOrSlug, 10, 64); err != nil {
			message = "Can't find post thread by slug: " + idOrSlug
		} else {
			message = "Can't find post thread by id: " + idOrSlug
		}
		resp := map[string]string{
			"message": message,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrUserNotFound) {
		resp := map[string]string{
			"message": "Can't find post author by nickname: ",
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrNoParent) {
		resp := map[string]string{
			"message": "Parent post was created in another thread",
		}
		httputils.RespondErr(ctx, http.StatusConflict, resp)
		return
	}
	if err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(ctx, http.StatusCreated, posts)
}

func (h *Handlers) ThreadInfo(ctx *fasthttp.RequestCtx) {
	idOrSlug := ctx.UserValue("slug_or_id").(string)
	thread, err := h.useCase.ThreadInfo(idOrSlug)
	if errors.Is(err, customErr.ErrForumNotFound) {
		resp := map[string]string{
			"message": "Can't find thread by slug or id: " + idOrSlug,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(ctx, http.StatusOK, thread)
}

func (h *Handlers) ChangeThread(ctx *fasthttp.RequestCtx) {
	var thread models.Thread
	if err := easyjson.Unmarshal(ctx.PostBody(), &thread); err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	idOrSlug := ctx.UserValue("slug_or_id").(string)
	thread, err := h.useCase.ChangeThread(idOrSlug, thread)

	if errors.Is(err, customErr.ErrThreadNotFound) {
		resp := map[string]string{
			"message": "Can't find thread by slug or id: " + idOrSlug,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(ctx, http.StatusOK, thread)
}

func (h *Handlers) GetPosts(ctx *fasthttp.RequestCtx) {
	idOrSlug := ctx.UserValue("slug_or_id").(string)

	// максимальное количество возвращаемых записей
	limit := int64(ctx.QueryArgs().GetUintOrZero("limit"))
	if limit == 0 {
		limit = 100
	}

	// Дата создания ветви обсуждения, с которой будут выводиться записи
	// (ветвь обсуждения с указанной датой попадает в результат выборки).
	since := int64(ctx.QueryArgs().GetUintOrZero("since"))

	// Вид сортировки:
	// flat - по дате, комментарии выводятся простым списком в порядке создания;
	// tree - древовидный, комментарии выводятся отсортированные в дереве
	// по N штук;
	// parent_tree - древовидные с пагинацией по родительским (parent_tree),
	// на странице N родительских комментов и все комментарии прикрепленные
	// к ним, в древвидном отображение.
	// Подробности: https://park.mail.ru/blog/topic/view/1191/
	//
	// Available values : flat, tree, parent_tree
	//
	// Default value : flat

	sort := string(ctx.QueryArgs().Peek("sort"))

	// Флаг сортировки по убыванию.
	desc := ctx.QueryArgs().GetBool("desc")

	var posts models.PostList
	var err error
	posts, err = h.useCase.GetPosts(idOrSlug, limit, since, sort, desc)

	if errors.Is(err, customErr.ErrThreadNotFound) {
		resp := map[string]string{
			"message": "Can't find thread by slug or id: " + idOrSlug,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(ctx, http.StatusOK, posts)
}

func (h *Handlers) VoteThread(ctx *fasthttp.RequestCtx) {
	var vote models.Vote
	if err := easyjson.Unmarshal(ctx.PostBody(), &vote); err != nil {
		httputils.Respond(ctx, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	idOrSlug := ctx.UserValue("slug_or_id").(string)
	nickname := vote.Nickname

	thread, err := h.useCase.VoteThread(idOrSlug, vote)

	if errors.Is(err, customErr.ErrThreadNotFound) {
		resp := map[string]string{
			"message": "Can't find thread by slug or id: " + idOrSlug,
		}
		httputils.RespondErr(ctx, http.StatusNotFound, resp)
		return
	}
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
	httputils.Respond(ctx, http.StatusOK, thread)
}
