package routes

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	zipkinUtil "meipian.cn/meigo/v2/util/zipkin"

	"chatgpt_server/controllers"
)

func pprofHandler(h http.HandlerFunc) gin.HandlerFunc {
	handler := http.HandlerFunc(h)
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func RouteInit(r *gin.Engine) {
	debugRoute := r.Group("/debug")
	{
		debugRoute.GET("/pprof/cmdline", pprofHandler(pprof.Cmdline))
		debugRoute.GET("/pprof/", pprofHandler(pprof.Index))
		debugRoute.GET("/pprof/profile", pprofHandler(pprof.Profile))
		debugRoute.GET("/pprof/symbol", pprofHandler(pprof.Symbol))
		debugRoute.GET("/pprof/trace", pprofHandler(pprof.Trace))
		// 404 page not found in /debug/pprof/allocs
		// --> https://github.com/gin-contrib/pprof/issues/15
		debugRoute.GET("/pprof/allocs", pprofHandler(pprof.Handler("allocs").ServeHTTP))
		debugRoute.GET("/pprof/block", pprofHandler(pprof.Handler("block").ServeHTTP))
		debugRoute.GET("/pprof/goroutine", pprofHandler(pprof.Handler("goroutine").ServeHTTP))
		debugRoute.GET("/pprof/heap", pprofHandler(pprof.Handler("heap").ServeHTTP))
		debugRoute.GET("/pprof/mutex", pprofHandler(pprof.Handler("mutex").ServeHTTP))
		debugRoute.GET("/pprof/threadcreate", pprofHandler(pprof.Handler("threadcreate").ServeHTTP))
	}

	var globalMiddleware = []gin.HandlerFunc{
		zipkinUtil.GinZipkinMiddleware,
	}

	root := r.Group("/", globalMiddleware...)

	chatCtrl := controllers.NewChat()
	chatRoute := root.Group("/chat")
	{
		chatRoute.POST("/sendMsg", chatCtrl.SendMsg)
	}
	chatGPTRoute := root.Group("/chatGPT")
	{
		chatGPTRoute.POST("/sendMsg", chatCtrl.SendChatGPTMsg)
	}
}
