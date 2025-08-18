package main

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type App struct {
	oRouter *mux.Router
	oDB     *sql.DB
	Router  *gin.Engine
	DB      *gorm.DB
	Log     *zerolog.Logger
}

func (a *App) InitialiseApp() {
	a.Router = gin.Default()
	//a.initialiseMiddleWare()
	a.initialiseRoutes()
	a.InitialiseDatabase()
	//a.PopulateDatabase()
}
/*
func (a *App) Initialize(host, user, password, dbname string) {

	a.oRouter = mux.NewRouter()
	a.initializeRoutes()

	connectionString :=
		fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)

	var err error
	a.oDB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Print("Connected to db successfully")
	}

}
 */

func (a *App) Run(port string) {
	a.Log.Info().Msgf("Server running on port [%s]", port)
	a.Log.Fatal().Err(a.Router.Run(port))
}
