package app

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"search_service/pkg/config"
	"search_service/pkg/repository"
)

type Application struct {
	Config *config.AppConfig
	Router *mux.Router
	ESRepo *repository.ElasticSearchRepository
}

func NewApplication(cfg *config.AppConfig, esRepo *repository.ElasticSearchRepository) *Application {
	app := &Application{
		Config: cfg,
		Router: mux.NewRouter(),
		ESRepo: esRepo,
	}
	app.setupRoutes()

	return app
}

func (a *Application) setupRoutes() {
	a.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to Go Event-Driven Search Service!")
	}).Methods("GET")
}

func (a *Application) StartApplication() {
	log.Printf("Starting HTTP server on port %s", a.Config.AppPort)
	log.Fatal(http.ListenAndServe(":"+a.Config.AppPort, a.Router))
}
