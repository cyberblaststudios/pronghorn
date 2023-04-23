package apiv1

import (
	"github.com/gin-gonic/gin"
)

func defaultRoute(context *gin.Context) {
	context.JSON(200, gin.H{
		"message": "helloword",
		"jj":      "hello",
	})
}

func Start(router *gin.Engine) {

	routerGroupV1 := router.Group("/apiv1")

	routerGroupV1.GET("/", defaultRoute)
}
