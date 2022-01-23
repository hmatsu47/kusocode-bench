package main

import (
	"github.com/gin-gonic/gin"
	"github.com/hmatsu47/kusocode-bench/controller"
)

func main() {
	router := gin.Default()
	router.POST("/start", controller.PostStart)
	router.Run(":5050")
}
