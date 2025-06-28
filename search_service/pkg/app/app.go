package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"search_service/pkg/config"
	"search_service/pkg/consumer"
	"search_service/pkg/handler"
	"search_service/pkg/repository"
	"search_service/pkg/service"
	"syscall"
)

type Application struct {
	Config      *config.AppConfig
	Router      *mux.Router
	ESRepo      *repository.ElasticSearchRepository
	NewsService *service.NewsService
	Consumer    *consumer.RabbitMQConsumer
}

func NewApplication(cfg *config.AppConfig, esRepo *repository.ElasticSearchRepository) *Application {
	app := &Application{
		Config: cfg,
		Router: mux.NewRouter(),
		ESRepo: esRepo,
	}
	adminHandler := handler.NewAdminHandler(app.ESRepo)
	newsService := service.NewNewsService(app.ESRepo)
	app.NewsService = newsService
	newsHandler := handler.NewNewsHandler(newsService)
	app.setupRoutes(adminHandler, newsHandler)

	rmqConsumer, err := consumer.NewRabbitMQConsumer(cfg, newsService)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ consumer: %v", err)
	}
	app.Consumer = rmqConsumer

	return app
}

func (a *Application) setupRoutes(
	adminHandler *handler.AdminHandler,
	newsHandler *handler.NewsHandler) {
	a.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to Go Event-Driven Search Service!")
	}).Methods("GET")

	adminRouter := a.Router.PathPrefix("/admin").Subrouter()
	adminRouter.HandleFunc("/info", adminHandler.GetElasticsearchInfo).Methods("GET")
	adminRouter.HandleFunc("/indices/info/{name}", adminHandler.GetExistElasticIndex).Methods("GET")
	adminRouter.HandleFunc("/indices/{name}", adminHandler.CreateIndex).Methods("POST")
	adminRouter.HandleFunc("/indices/{name}", adminHandler.DeleteIndex).Methods("DELETE")

	newsRouter := a.Router.PathPrefix("/news").Subrouter()
	newsRouter.HandleFunc("", newsHandler.SearchNews).Methods("GET")
	newsRouter.HandleFunc("/{id}", newsHandler.GetNewsByID).Methods("GET")

}

func (a *Application) StartApplication() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		a.Consumer.StartConsuming(ctx)
	}()

	log.Printf("Starting HTTP server on port %s", a.Config.AppPort)
	err := http.ListenAndServe(":"+a.Config.AppPort, a.Router)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server failed: %v", err)
	}

	<-ctx.Done()
	log.Println("Shutting down application gracefully...")
	a.Consumer.Close()
	log.Println("RabbitMQ Consumer closed during shutdown.")
}
