package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func bbLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		// 给Context实例设置一个值
		c.Set("geektutu", "1111")
		// 请求前
		c.Next()
		// 请求后
		latency := time.Since(t)
		log.Print(latency)
	}
}

// 全局统一出错处理
type GlobalErrorHandlerFunc func(*gin.Context) error

func wrapper(handler GlobalErrorHandlerFunc) func(c *gin.Context) {
	return func(c *gin.Context) {
		err := handler(c)
		if err != nil {
			var apiException *APIException
			if h, ok := err.(*APIException); ok {
				apiException = h
			} else {
				if gin.Mode() == "debug" {
					apiException = UnknownError(err.Error())
				} else {
					apiException = ServerError()
				}
			}
			apiException.Request = c.Request.Method + " to " + c.Request.URL.String()
			c.JSON(apiException.Code, apiException)
			return
		}
	}
}

func main() {

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Who are you?")
	})
	r.GET("ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// 路径参数
	r.GET("/user/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		ctx.String(http.StatusOK, "hello %s", name)
	})
	// Query参数
	r.GET("/user/query", func(ctx *gin.Context) {
		name := ctx.Query("name")
		role := ctx.DefaultQuery("role", "teacher")
		ctx.String(http.StatusOK, "%s is a %s", name, role)
	})
	// 表单提交
	r.POST("/form", func(ctx *gin.Context) {
		uname := ctx.PostForm("uname")
		pword := ctx.DefaultPostForm("pword", "999999")
		ctx.JSON(http.StatusOK, gin.H{
			"username": uname,
			"password": pword,
		})
	})
	// 表单提交及Query混合
	r.POST("/posts", func(ctx *gin.Context) {
		id := ctx.Query("id")
		page := ctx.DefaultQuery("page", "1")
		uname := ctx.PostForm("uname")
		pword := ctx.DefaultPostForm("pword", "999999")
		ctx.JSON(http.StatusOK, gin.H{
			"id":       id,
			"page":     page,
			"username": uname,
			"password": pword,
		})
	})
	//路由分组，比如某几个路由都在 /api/user/v1 下面
	defaultHandler := func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"path": ctx.FullPath(),
		})
	}
	v1 := r.Group("/v1")
	{
		v1.GET("/posts", defaultHandler)
		v1.GET("/series", defaultHandler)
	}
	v2 := r.Group("/v2")
	{
		v2.GET("/posts", defaultHandler)
		v2.GET("/series", defaultHandler)
	}
	// 上传文件
	r.POST("upload", func(ctx *gin.Context) {
		file, _ := ctx.FormFile("file")
		ctx.String(http.StatusOK, "%s uploaded!", file.Filename)
	})
	// 中间件，全局
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	// 中间件只在只定路由发生作用
	r.GET("/single", bbLogger(), func(ctx *gin.Context) {
		ctx.String(http.StatusOK, ctx.GetString("geektutu"))
	})

	r.GET("/global/error", wrapper(gerror))
	r.Run()
}

func gerror(c *gin.Context) error {
	name := c.Query("name")
	if name == "abcd" {
		return newAPIException(http.StatusBadRequest, "param error")
	}
	c.JSON(200, gin.H{
		"message": name,
	})
	return nil
}

type APIException struct {
	Code    int    `json:"-"`
	Msg     string `json:"msg"`
	Request string `json:"request"`
}

func (e *APIException) Error() string {
	return e.Msg
}

func newAPIException(code int, msg string) *APIException {
	return &APIException{
		Code: code,
		Msg:  msg,
	}
}

func ServerError() *APIException {
	return newAPIException(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
}

func NotFound() *APIException {
	return newAPIException(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func UnknownError(message string) *APIException {
	return newAPIException(http.StatusForbidden, message)
}
