package handlers

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/httputils"
	"DBForum/internal/app/models"
	postUseCase "DBForum/internal/app/post/usecase"
	"errors"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
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

func (h *Handlers) GetInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.Println(err)
		httputils.Respond(w, http.StatusInternalServerError, nil)
		return
	}
	// Включение полной информации о соответвующем объекте сообщения.
	// Если тип объекта не указан, то полная информация об этих объектах не
	// передаётся.
	// values: user/forum/thread
	related := strings.Split(r.URL.Query().Get("related"), ",")

	postInfo, err := h.useCase.GetPostInfoByID(id, related)

	if errors.Is(err, customErr.ErrPostNotFound) {
		resp := map[string]string{
			"message": "Can't find post with id: ",
		}
		httputils.RespondErr(w, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		log.Println(err)
		httputils.Respond(w, http.StatusInternalServerError, nil)
		return
	}
	httputils.Respond(w, http.StatusOK, postInfo)
}

func (h *Handlers) ChangeMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	post := &models.Post{}
	if err := easyjson.UnmarshalFromReader(r.Body, post); err != nil {
		log.Println(err)
		httputils.Respond(w, http.StatusInternalServerError, post)
		return
	}
	params := mux.Vars(r)
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.Println(err)
		httputils.Respond(w, http.StatusInternalServerError, post)
		return
	}
	post.ID = id
	post, err = h.useCase.ChangeMessage(*post)
	if errors.Is(err, customErr.ErrPostNotFound) {
		resp := map[string]string{
			"message": "Can't find post with id: " + strconv.FormatUint(id, 10),
		}
		httputils.RespondErr(w, http.StatusNotFound, resp)
		return
	}
	if err != nil {
		log.Println(err)
		httputils.Respond(w, http.StatusInternalServerError, post)
		return
	}
	httputils.Respond(w, http.StatusOK, post)
}
