package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"net/http"
	"bytes"
)

func main() {
	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/get", getForward)
	router.Run(":8090")
}

func getForward(ctx *gin.Context) {
	url, ok := ctx.GetQuery("url")
	if !ok {
		return
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    "FAIL",
			"message": "forward: " + err.Error(),
			"data":    nil,
		})
		return
	}
	request.Header = ctx.Request.Header
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    "FAIL",
			"message": "forward: " + err.Error(),
			"data":    nil,
		})
		return
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	s := buf.String()
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(200, "%s", s)
	return
}
