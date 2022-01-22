package routes

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vivek-yadav/rabbit/zlog"
)

func Routes(router *gin.Engine) {
	router.GET("/", welcome)
	router.GET("/error", errResp)
}
func welcome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "Welcome To API",
		"time":    fmt.Sprint(time.Now().Unix()),
	})
	// time.Sleep(time.Second)
	return
}
func errResp(c *gin.Context) {

	tryError(c)
	c.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": "Welcome To API",
		"time":    fmt.Sprint(time.Now().Unix()),
	})
	return
}

func tryError(c *gin.Context) (err error) {
	// defer catch(err)
	err = errors.New("Faild to understand you, speak clearly.")
	zlog.CheckAndAbortAPIError(err, c, http.StatusBadRequest)
	fmt.Println("NO ERROR")
	return
}

func catch(err error) {
	if err := recover(); err != nil {

	}
}
