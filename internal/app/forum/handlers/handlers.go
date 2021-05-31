package handlers

import (
	customErr "DBForum/internal/app/errors"
	forumUseCase "DBForum/internal/app/forum/usecase"
	"DBForum/internal/app/httputils"
	"DBForum/internal/app/models"
	"errors"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"log"
	"net/http"
	"strconv"
)

type Handlers struct {
	useCase forumUseCase.UseCase
}

func NewHandler(useCase forumUseCase.UseCase) *Handlers {
	return &Handlers{
		useCase: useCase,
	}
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	forum := &models.Forum{}

	if err := easyjson.UnmarshalFromReader(r.Body, forum); err != nil {
		log.Println(err)
		httputils.Respond(w, http.StatusInternalServerError, nil)
		return
	}

	var err error
	nickname := forum.User
	forum, err = h.useCase.CreateForum(forum)
	if errors.Is(err, customErr.ErrUserNotFound) {
		resp := map[string]string{
			"message": "Can't find user with nickname: " + nickname,
		}
		httputils.RespondErr(w, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrDuplicate) {
		httputils.Respond(w, http.StatusConflict, forum)
		return
	}
	if err != nil {
		log.Println(err)
		httputils.Respond(w, http.StatusInternalServerError, nil)
		return
	}
	httputils.Respond(w, http.StatusCreated, forum)
}

func (h *Handlers) Details(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	slug := params["slug"]
	forum, err := h.useCase.GetInfoBySlug(slug)
	if errors.Is(err, customErr.ErrForumNotFound) {
		resp := map[string]string{
			"message": "Can't find forum with slug: " + slug,
		}
		httputils.RespondErr(w, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusOK, forum)
}

func (h *Handlers) CreateThread(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	thread := &models.Thread{}
	if err := easyjson.UnmarshalFromReader(r.Body, thread); err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	params := mux.Vars(r)
	forumSlug := params["slug"]
	nickname := thread.Author
	thread.Forum = forumSlug

	var err error
	thread, err = h.useCase.CreateThread(thread)
	if errors.Is(err, customErr.ErrUserNotFound) {
		resp := map[string]string{
			"message":"Can't find thread author by nickname: " + nickname,
		}
		httputils.RespondErr(w, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrForumNotFound) {
		resp := map[string]string{
			"message": "Can't find thread forum by slug: " + forumSlug,
		}
		httputils.RespondErr(w, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrDuplicate) {
		httputils.Respond(w, http.StatusConflict, thread)
		return
	}
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusCreated, thread)
}

func (h *Handlers) GetUsers(w http.ResponseWriter, r *http.Request) {
	forumSlug := mux.Vars(r)["slug"]
	// максимальное количество возвращаемых записей
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	// Идентификатор пользователя, с которого будут выводиться пользоватли
	//(пользователь с данным идентификатором в результат не попадает).
	since := r.URL.Query().Get("since")
	// Флаг сортировки по убыванию.
	desc, _ := strconv.ParseBool(r.URL.Query().Get("desc"))

	var users models.UserList
	var err error
	users, err = h.useCase.GetForumUsers(forumSlug, limit, since, desc)
	if errors.Is(err, customErr.ErrForumNotFound) {
		resp := map[string]string{
			"message": "Can't find forum by slug: " + forumSlug,
		}
		httputils.RespondErr(w, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	httputils.Respond(w, http.StatusOK, users)
}

func (h *Handlers) GetThreads(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	forumSlug := params["slug"]
	var threads models.ThreadList

	// максимальное количество возвращаемых записей
	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	// Дата создания ветви обсуждения, с которой будут выводиться записи
	// (ветвь обсуждения с указанной датой попадает в результат выборки).
	since := r.URL.Query().Get("since")
	// Флаг сортировки по убыванию.
	desc, _ := strconv.ParseBool(r.URL.Query().Get("desc"))

	var err error
	threads, err = h.useCase.GetForumThreads(forumSlug, limit, since, desc)
	if errors.Is(err, customErr.ErrForumNotFound) {
		resp := map[string]string{
			"message": "Can't find forum by slug: " + forumSlug,
		}
		httputils.RespondErr(w, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusOK, threads)
}
