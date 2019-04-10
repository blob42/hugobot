package main

import (
	"hugobot/bitcoin"
	"hugobot/config"
	"hugobot/feeds"
	"io"
	"os"
	"strconv"

	gum "git.sp4ke.com/sp4ke/gum.git"
	"github.com/gin-gonic/gin"
)

var (
	apiLogFile *os.File
)

type API struct {
	router *gin.Engine
}

func (api *API) Run(m gum.UnitManager) {

	feedsRoute := api.router.Group("/feeds")
	{
		feedCtrl := &feeds.FeedCtrl{}

		feedsRoute.POST("/", feedCtrl.Create)
		feedsRoute.DELETE("/:id", feedCtrl.Delete)
		feedsRoute.GET("/", feedCtrl.List) // Get all
		//feedsRoute.Get("/:id", feedCtrl.GetById) // Get one
	}

	btcRoute := api.router.Group("/btc")
	{
		btcRoute.GET("/address", bitcoin.GetAddressCtrl)
	}

	// Run router
	go func() {

		err := api.router.Run(":" + strconv.Itoa(config.C.ApiPort))
		if err != nil {
			panic(err)
		}
	}()

	// Wait for stop signal
	<-m.ShouldStop()

	// Shutdown
	api.Shutdown()
	m.Done()
}

func (api *API) Shutdown() {}

func NewApi() *API {
	apiLogFile, _ = os.Create(".api.log")
	gin.DefaultWriter = io.MultiWriter(apiLogFile, os.Stdout)

	api := &API{
		router: gin.Default(),
	}

	return api
}
