package main

import (
	"github.com/gin-gonic/gin"
)

func (a *App) initialiseMiddleWare() {
	a.Log.Debug().Msg("Initialising middleware")
	a.Router.Use(a.LoggingMiddleware())
	a.Router.Use(gin.Recovery())
	//a.Router.Use(a.auditMiddleware())
}

//-----------------------------------------------------------------------------

func (a *App) LoggingMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		a.Log.Debug().Msgf("Route [%s]; Method [%s]; IP [%s]", c.Request.URL.Path, c.Request.Method, c.Request.RemoteAddr)
		c.Next()
	}
}