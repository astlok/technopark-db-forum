package handlers

import (
	customErr "DBForum/internal/app/errors"
	"DBForum/internal/app/httputils"
	"DBForum/internal/app/models"
	userUseCase "DBForum/internal/app/user/usecase"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
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

func(h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	params := mux.Vars(r)
	nickname := params["nickname"]

	user := models.User{Nickname: nickname}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	err := h.useCase.CreateUser(user)
	if errors.Is(err, customErr.ErrDuplicate) {
		users, err := h.useCase.GetUsersByNickAndEmail(user.Nickname, user.Email)
		if err != nil {
			httputils.Respond(w, http.StatusInternalServerError, nil)
			log.Println(err)
			return
		}
		httputils.Respond(w, http.StatusConflict, users)
		return
	}
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusCreated, user)
}

func(h *Handlers) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	nickname := params["nickname"]
	user := &models.User{Nickname: nickname}

	user, err := h.useCase.GetUserInfo(nickname)

	if errors.Is(err, customErr.ErrUserNotFound)  {
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

	httputils.Respond(w, http.StatusOK, user)
}

func(h *Handlers) ChangeUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	params := mux.Vars(r)
	nickname := params["nickname"]

	user := models.User{Nickname: nickname}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}

	err := h.useCase.ChangeUser(&user)
	if errors.Is(err, customErr.ErrUserNotFound) {
		resp := map[string]string{
			"message":  "Can't find user by nickname: " + nickname,
		}
		httputils.Respond(w, http.StatusNotFound, resp)
		return
	}
	if errors.Is(err, customErr.ErrConflict) {
		userNick, err := h.useCase.GetUserNickByEmail(user.Email)
		if err != nil {
			httputils.Respond(w, http.StatusInternalServerError, nil)
			log.Println(err)
			return
		}
		resp := map[string]string{
			"message": "This email is already registered by user: " + userNick,
		}
		httputils.Respond(w, http.StatusConflict, resp)
		return
	}
	httputils.Respond(w, http.StatusOK, user)
}
