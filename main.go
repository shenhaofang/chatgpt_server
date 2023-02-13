package main

import (
	"fmt"
	"net/http"
	"os"

	"meipian.cn/meigo/v2/util"

	"meipian.cn/meigo/v2/config"
	"meipian.cn/meigo/v2/log"
	zipkinUtil "meipian.cn/meigo/v2/util/zipkin"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/urfave/cli/v2"

	"chatgpt_server/repos"
	"chatgpt_server/routes"
)

func startListen() {
	engin := util.NewGin()
	routes.RouteInit(engin)

	addr := ":" + config.GetDft("port", "10100")
	s := &http.Server{
		Addr:    addr,
		Handler: engin,
	}
	fmt.Println("Server listen on", addr)
	err := gracehttp.Serve(s)
	if err != nil {
		fmt.Println(err)
		log.Err(err.Error())
	}
}

func main() {
	zipkinUtil.InitZipkinWithApolloConfig()
	repos.InitChatGPTs()
	app := cli.NewApp()
	app.Name = "user_feed go server"
	app.Action = func(c *cli.Context) error {
		fmt.Println("Run Http Server")
		startListen()
		return nil
	}
	app.Run(os.Args)
}
