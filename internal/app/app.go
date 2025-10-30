package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lzimin05/course-todo/config"
	"github.com/lzimin05/course-todo/internal/infrastructure/redis"
	"github.com/lzimin05/course-todo/internal/infrastructure/repository"
	authrepo "github.com/lzimin05/course-todo/internal/infrastructure/repository/auth"
	autht "github.com/lzimin05/course-todo/internal/transport/auth"
	"github.com/lzimin05/course-todo/internal/transport/jwt"
	"github.com/lzimin05/course-todo/internal/transport/middleware"
	authuc "github.com/lzimin05/course-todo/internal/usecase/auth"
	"github.com/sirupsen/logrus"

	userrepo "github.com/lzimin05/course-todo/internal/infrastructure/repository/user"
	usert "github.com/lzimin05/course-todo/internal/transport/user"
	useruc "github.com/lzimin05/course-todo/internal/usecase/user"

	taskRepo "github.com/lzimin05/course-todo/internal/infrastructure/repository/task"
	taskt "github.com/lzimin05/course-todo/internal/transport/task"
	taskuc "github.com/lzimin05/course-todo/internal/usecase/task"

	noteRepo "github.com/lzimin05/course-todo/internal/infrastructure/repository/note"
	notet "github.com/lzimin05/course-todo/internal/transport/note"
	noteuc "github.com/lzimin05/course-todo/internal/usecase/note"
)

// App объединяет все компоненты приложения
type App struct {
	conf   *config.Config
	logger *logrus.Logger
	db     *sql.DB
	router *mux.Router
}

// NewApp инициализирует приложение
func NewApp(conf *config.Config) (*App, error) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Подключение к БД
	dbConnStr, err := repository.GetConnectionString(conf.DBConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	config.ConfigureDB(db, conf.DBConfig)

	redisAuthClient, err := redis.NewClient(conf.AuthRedisConfig)
	if err != nil {
		logger.Fatalf("redis auth connection error: %v", err)
	}
	authRepo := authrepo.New(db)
	redisAuthRepo := redis.NewAuthRepository(redisAuthClient, conf.JWTConfig)
	tokenator := jwt.NewTokenator(conf.JWTConfig)
	authUC := authuc.New(authRepo, tokenator, redisAuthRepo)
	authHandler := autht.New(authUC, conf)

	userRepo := userrepo.New(db)
	userUC := useruc.New(userRepo)
	userHandler := usert.New(userUC, conf)

	noteRepo := noteRepo.NewNoteRepository(db)
	noteUC := noteuc.NewNoteUsecase(noteRepo)
	noteHandler := notet.NewNoteHandler(noteUC, conf)

	// Настройка маршрутизатора
	router := mux.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return middleware.LogRequest(logger, next)
	})

	apiRouter := router.PathPrefix("/api").Subrouter()

	authRouter := apiRouter.PathPrefix("/auth").Subrouter()
	{
		authRouter.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)
		authRouter.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
		authRouter.Handle("/logout",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(authHandler.Logout)),
		).Methods(http.MethodPost)
	}

	userRouter := apiRouter.PathPrefix("/users").Subrouter()
	{
		userRouter.Handle("/me",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(userHandler.GetMe)),
		).Methods(http.MethodGet)
	}

	taskRepository := taskRepo.New(db)
	taskUseCase := taskuc.New(taskRepository)
	taskHandler := taskt.New(taskUseCase, conf)

	taskRouter := apiRouter.PathPrefix("/todo").Subrouter()
	{
		//taskRouter.HandleFunc("/create", taskHandler.CreateTask).Methods(http.MethodPost)
		taskRouter.Handle("/create",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(taskHandler.CreateTask)),
		).Methods(http.MethodPost)
		taskRouter.Handle("/all",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(taskHandler.GetTasksByUserID)),
		).Methods(http.MethodGet)
		taskRouter.Handle("/{taskId}/edit",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(taskHandler.UpdateTask)),
		).Methods(http.MethodPut)
		taskRouter.Handle("/{taskId}/edit",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(taskHandler.UpdateTaskStatus)),
		).Methods(http.MethodPatch)
		taskRouter.Handle("/{taskId}/",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(taskHandler.DeleteTask)),
		).Methods(http.MethodDelete)
	}

	noteRouter := apiRouter.PathPrefix("/notes").Subrouter()
	{
		noteRouter.Handle("/all",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(noteHandler.GetAllNotes)),
		).Methods(http.MethodGet)

		noteRouter.Handle("/create",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(noteHandler.CreateNote)),
		).Methods(http.MethodPost)

		noteRouter.Handle("/{noteId}/edit",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(noteHandler.UpdateNote)),
		).Methods(http.MethodPut)

		noteRouter.Handle("/{noteId}",
			middleware.AuthMiddleware(tokenator)(http.HandlerFunc(noteHandler.DeleteNote)),
		).Methods(http.MethodDelete)
	}

	return &App{
		conf:   conf,
		logger: logger,
		db:     db,
		router: router,
	}, nil
}

// Run запускает HTTP-сервер
func (a *App) Run() {
	server := &http.Server{
		Addr:    ":" + a.conf.ServerConfig.Port,
		Handler: a.router,
	}

	a.logger.Infof("Starting server on port %s", a.conf.ServerConfig.Port)
	if err := server.ListenAndServe(); err != nil {
		a.logger.Fatalf("Server failed: %v", err)
	}
}

func (a *App) GetRouter() *mux.Router {
	return a.router
}
