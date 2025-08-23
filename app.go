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
	ORouter *mux.Router
	ODB     *sql.DB
	Router  *gin.Engine
	DB      *gorm.DB
	Log     *zerolog.Logger
}

func (a *App) InitialiseApp() {
	a.Router = gin.Default()
	a.InitialiseRoutes()
	a.InitialiseDatabase()
}

func (a *App) Run(port string) {
	a.Log.Info().Msgf("Server running on port [%s]", port)
	a.Log.Fatal().Err(a.Router.Run(port))
}
