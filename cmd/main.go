package main

import (
	"DBForum/internal/app/database"
	forumHandlers "DBForum/internal/app/forum/handlers"
	forumRepo "DBForum/internal/app/forum/repository"
	forumUCase "DBForum/internal/app/forum/usecase"
	postHandlers "DBForum/internal/app/post/handlers"
	postRepo "DBForum/internal/app/post/repository"
	postUCase "DBForum/internal/app/post/usecase"
	serviceHandlers "DBForum/internal/app/service/handlers"
	serviceRepo "DBForum/internal/app/service/repository"
	serviceUCase "DBForum/internal/app/service/usecase"
	"fmt"
	"github.com/sirupsen/logrus"

	threadHandlers "DBForum/internal/app/thread/handlers"
	threadRepo "DBForum/internal/app/thread/repository"
	threadUCase "DBForum/internal/app/thread/usecase"

	userHandlers "DBForum/internal/app/user/handlers"
	userRepo "DBForum/internal/app/user/repository"
	userUCase "DBForum/internal/app/user/usecase"

	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {

	postgres, err := database.NewPostgres()

	if err != nil {
		log.Fatal(err)
	}

	forumRepository := forumRepo.NewRepo(postgres.GetPostgres())
	if err := forumRepository.Prepare(); err != nil {
		log.Fatalln(err)
	}
	postRepository := postRepo.NewRepo(postgres.GetPostgres())
	if err := postRepository.Prepare(); err != nil {
		log.Fatalln(err)
	}
	serviceRepository := serviceRepo.NewRepo(postgres.GetPostgres())
	threadRepository := threadRepo.NewRepo(postgres.GetPostgres())
	if err := threadRepository.Prepare(); err != nil {
		log.Fatalln(err)
	}
	userRepository := userRepo.NewRepo(postgres.GetPostgres())
	if err := userRepository.Prepare(); err != nil {
		log.Fatalln(err)
	}

	forumUseCase := forumUCase.NewUseCase(*forumRepository, *userRepository, *threadRepository)
	postUseCase := postUCase.NewUseCase(*postRepository, *userRepository, *threadRepository, *forumRepository)
	serviceUseCase := serviceUCase.NewUseCase(*serviceRepository)
	threadUseCase := threadUCase.NewUseCase(*threadRepository, *postRepository)
	userUseCase := userUCase.NewUseCase(*userRepository)

	forumHandler := forumHandlers.NewHandler(*forumUseCase)
	postHandler := postHandlers.NewHandler(*postUseCase)
	serviceHandler := serviceHandlers.NewHandler(*serviceUseCase)
	threadHandler := threadHandlers.NewHandler(*threadUseCase)
	userHandler := userHandlers.NewHandler(*userUseCase)

	router := mux.NewRouter()

	router.Use(commonMiddleware)
	//router.Use(LoggingRequest)
	forum := router.PathPrefix("/api/forum").Subrouter()

	//done
	forum.HandleFunc("/create", forumHandler.Create).Methods(http.MethodPost)

	forum.HandleFunc("/{slug}/details", forumHandler.Details).Methods(http.MethodGet)

	//done
	forum.HandleFunc("/{slug}/create", forumHandler.CreateThread).Methods(http.MethodPost)

	forum.HandleFunc("/{slug}/users", forumHandler.GetUsers).Methods(http.MethodGet)

	forum.HandleFunc("/{slug}/threads", forumHandler.GetThreads).Methods(http.MethodGet)

	post := router.PathPrefix("/api/post").Subrouter()

	post.HandleFunc("/{id:[0-9]+}/details", postHandler.GetInfo).Methods(http.MethodGet)

	//done
	post.HandleFunc("/{id:[0-9]+}/details", postHandler.ChangeMessage).Methods(http.MethodPost)

	service := router.PathPrefix("/api/service").Subrouter()

	service.HandleFunc("/clear", serviceHandler.ClearDB).Methods(http.MethodPost)

	service.HandleFunc("/status", serviceHandler.Status).Methods(http.MethodGet)

	thread := router.PathPrefix("/api/thread").Subrouter()

	//done
	thread.HandleFunc("/{slug_or_id}/create", threadHandler.CreatePost).Methods(http.MethodPost)

	thread.HandleFunc("/{slug_or_id}/details", threadHandler.ThreadInfo).Methods(http.MethodGet)


	//done
	thread.HandleFunc("/{slug_or_id}/details", threadHandler.ChangeThread).Methods(http.MethodPost)

	thread.HandleFunc("/{slug_or_id}/posts", threadHandler.GetPosts).Methods(http.MethodGet)

	//done
	thread.HandleFunc("/{slug_or_id}/vote", threadHandler.VoteThread).Methods(http.MethodPost)

	user := router.PathPrefix("/api/user").Subrouter()

	//done
	user.HandleFunc("/{nickname}/create", userHandler.CreateUser).Methods(http.MethodPost)

	user.HandleFunc("/{nickname}/profile", userHandler.GetUserInfo).Methods(http.MethodGet)

	//done
	user.HandleFunc("/{nickname}/profile", userHandler.ChangeUser).Methods(http.MethodPost)

	server := &http.Server{
		Handler: router,
		Addr:    ":5000",
	}

	fmt.Printf("Starting server on port %s\n", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func LoggingRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"url":    r.URL,
			"method": r.Method,
			"body":   r.Body,
		}).Info()
		next.ServeHTTP(w, r)
	})
}
