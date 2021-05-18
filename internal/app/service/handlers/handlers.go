package handlers

import (
	"DBForum/internal/app/httputils"
	serviceUseCase "DBForum/internal/app/service/usecase"
	"log"
	"net/http"
)

type Handlers struct {
	useCase serviceUseCase.UseCase
}

func NewHandler(useCase serviceUseCase.UseCase) *Handlers {
	return &Handlers{
		useCase: useCase,
	}
}

func(h *Handlers) ClearDB(w http.ResponseWriter, r *http.Request) {
	err := h.useCase.ClearDB()
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusOK, nil)
}

func(h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	numRec, err := h.useCase.Status()
	if err != nil {
		httputils.Respond(w, http.StatusInternalServerError, nil)
		log.Println(err)
		return
	}
	httputils.Respond(w, http.StatusOK, numRec)
}
