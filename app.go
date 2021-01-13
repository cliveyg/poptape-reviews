package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(host, user, password, dbname string) {

	a.Router = mux.NewRouter()
	a.initializeRoutes()

	connectionString :=
		fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Print("Connected to db successfully")
	}

}

func (a *App) Run(addr string) {
	log.Print(fmt.Sprintf("Server running on port [%s]", addr))
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
