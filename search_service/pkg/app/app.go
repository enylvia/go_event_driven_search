package app

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"search_service/pkg/config"
	"search_service/pkg/handler"
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
	adminHandler := handler.NewAdminHandler(app.ESRepo)
	app.setupRoutes(adminHandler)

	return app
}

func (a *Application) setupRoutes(adminHandler *handler.AdminHandler) {
	a.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to Go Event-Driven Search Service!")
	}).Methods("GET")

	adminRouter := a.Router.PathPrefix("/admin").Subrouter()
	adminRouter.HandleFunc("/info", adminHandler.GetElasticsearchInfo).Methods("GET")
	adminRouter.HandleFunc("/indices/info", adminHandler.GetExistElasticIndex).Methods("GET")
	adminRouter.HandleFunc("/indices/{name}", adminHandler.CreateIndex).Methods("POST")
	adminRouter.HandleFunc("/indices/{name}", adminHandler.DeleteIndex).Methods("DELETE")

}

func (a *Application) StartApplication() {
	log.Printf("Starting HTTP server on port %s", a.Config.AppPort)
	log.Fatal(http.ListenAndServe(":"+a.Config.AppPort, a.Router))
}
