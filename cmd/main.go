package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/dmitryovchinnikov/mvc/internal/app"
	"github.com/dmitryovchinnikov/mvc/internal/app/controllers"
	"github.com/dmitryovchinnikov/mvc/internal/app/models"
	"github.com/gorilla/mux"
)

func main() {
	var err error

	boolPtr := flag.Bool("prod", false, "Provide this flag in production. This ensures that a .config file is provided before the application starts.")
	flag.Parse()

	cfg := app.LoadConfig(*boolPtr)
	dbCfg := cfg.Database
	services, err := models.NewServices(
		models.WithGORMLogMode(dbCfg.ConnectionInfo(), !cfg.IsProd()),
		models.WithTable(),
	)
	must(err)
	defer func(services *models.Services) {
		err := services.Close()
		if err != nil {
			panic(err)
		}
	}(services)
	err = services.AutoMigrate()
	if err != nil {
		return
	}

	r := mux.NewRouter()
	tblC := controllers.NewTables(services.Table, r)

	// Assets
	assetHandler := http.FileServer(http.Dir("./assets/"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)

	// Table routes
	r.Handle("/tables", http.HandlerFunc(tblC.Index)).Methods("GET")
	r.Handle("/tables/new", http.Handler(tblC.New)).Methods("GET")
	r.HandleFunc("/tables", http.HandlerFunc(tblC.Create)).Methods("POST")
	r.HandleFunc("/tables/edit", http.HandlerFunc(tblC.Edit)).Methods("GET").Name(controllers.EditTable)
	r.HandleFunc("/tables/update", http.HandlerFunc(tblC.Update)).Methods("POST")
	r.HandleFunc("/tables/text", http.HandlerFunc(tblC.TextUpload)).Methods("POST")

	r.HandleFunc("/tables/", tblC.Show).Methods("GET").Name(controllers.ShowTable)

	fmt.Printf("Starting the server on :%d...\n", cfg.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r)
	if err != nil {
		return
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
