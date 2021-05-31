package handlers

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/httputils"
	"DBForum/internal/app/models"
	threadUseCase "DBForum/internal/app/thread/usecase"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Handlers struct {
	useCase threadUseCase.UseCase
}

func NewHandler(useCase threadUseCase.UseCase) *Handlers {
	return &Handlers{
		useCase: useCase,
	}
}

func (h *Handlers) CreatePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var posts []models.Post
	if err := json.NewDecoder(r.Body).Decode(&posts); err != nil {
		httputils.Respond(w, http.StatusInternalServerError, posts)
		log.Println(err)
		return
	}
	params := mux.Vars(r)
	idOrSlug := params["slug_or_id"]
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
		httputils.Respond(w, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrUserNotFound) {
		nick := strings.Split(errors.Unwrap(err).Error(), ":")
		resp := map[string]string{
			"message": "Can't find post author by nickname: " + nick[0],
		}
		httputils.Respond(w, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrNoParent) {
		resp := map[string]string{
			"message": "Parent post was created in another thread",
		}
		httputils.Respond(w, http.StatusConflict, resp)
		return
	}
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusCreated, posts)
}

func (h *Handlers) ThreadInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	idOrSlug := params["slug_or_id"]
	thread, err := h.useCase.ThreadInfo(idOrSlug)
	if errors.Is(err, customErr.ErrForumNotFound) {
		resp := map[string]string{
			"message": "Can't find thread by slug or id: " + idOrSlug,
		}
		httputils.Respond(w, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusOK, thread)
}

func (h *Handlers) ChangeThread(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var thread models.Thread
	if err := json.NewDecoder(r.Body).Decode(&thread); err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	params := mux.Vars(r)
	idOrSlug := params["slug_or_id"]
	thread, err := h.useCase.ChangeThread(idOrSlug, thread)

	if errors.Is(err, customErr.ErrThreadNotFound) {
		resp := map[string]string{
			"message": "Can't find thread by slug or id: " + idOrSlug,
		}
		httputils.Respond(w, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusOK, thread)
}

func (h *Handlers) GetPosts(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	idOrSlug := params["slug_or_id"]

	// максимальное количество возвращаемых записей
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if limit == 0 {
		limit = 100
	}

	// Дата создания ветви обсуждения, с которой будут выводиться записи
	// (ветвь обсуждения с указанной датой попадает в результат выборки).
	since, _ := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)

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

	sort := r.URL.Query().Get("sort")

	// Флаг сортировки по убыванию.
	desc, err := strconv.ParseBool(r.URL.Query().Get("desc"))

	posts, err := h.useCase.GetPosts(idOrSlug, limit, since, sort, desc)

	if errors.Is(err, customErr.ErrThreadNotFound) {
		resp := map[string]string{
			"message": "Can't find thread by slug or id: " + idOrSlug,
		}
		httputils.Respond(w, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusOK, posts)
}

func (h *Handlers) VoteThread(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var vote models.Vote
	if err := json.NewDecoder(r.Body).Decode(&vote); err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	params := mux.Vars(r)
	idOrSlug := params["slug_or_id"]
	nickname := vote.Nickname

	thread, err := h.useCase.VoteThread(idOrSlug, vote)

	if errors.Is(err, customErr.ErrThreadNotFound) {
		resp := map[string]string{
			"message": "Can't find thread by slug or id: " + idOrSlug,
		}
		httputils.Respond(w, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrUserNotFound) {
		resp := map[string]string{
			"message": "Can't find user by nickname: " + nickname,
		}
		httputils.Respond(w, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusOK, thread)
}
