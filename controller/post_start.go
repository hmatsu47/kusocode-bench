package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hmatsu47/kusocode-bench/apimodel"
	"github.com/hmatsu47/kusocode-bench/service"
)

func PostStart(c *gin.Context) {
	var ipAddress string
	var result apimodel.Result

	ipAddress = c.ClientIP()
	// ipAddress = "127.0.0.1"
	result = service.Bench(ipAddress)

	c.JSON(http.StatusOK, gin.H{"result": result})
}
