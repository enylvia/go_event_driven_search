package app

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"search_service/pkg/config"
	"search_service/pkg/handler"
	"search_service/pkg/repository"
	"search_service/pkg/service"
)

type Application struct {
	Config      *config.AppConfig
	Router      *mux.Router
	ESRepo      *repository.ElasticSearchRepository
	NewsService *service.NewsService
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
	log.Printf("Starting HTTP server on port %s", a.Config.AppPort)
	log.Fatal(http.ListenAndServe(":"+a.Config.AppPort, a.Router))
}
