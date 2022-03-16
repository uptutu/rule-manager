package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/kit/transport"
	"github.com/tkeel-io/rule-manager/pkg/server"
	"github.com/tkeel-io/rule-manager/pkg/service"

	// User import.
	helloworld "github.com/tkeel-io/rule-manager/api/helloworld/v1"

	openapi "github.com/tkeel-io/rule-manager/api/openapi/v1"
	Rules_v1 "github.com/tkeel-io/rule-manager/api/rule/v1"
)

var (
	// Name app.
	Name string
	// HTTPAddr string.
	HTTPAddr string
	// GRPCAddr string.
	GRPCAddr string
)

func init() {
	flag.StringVar(&Name, "name", "hello", "app name.")
	flag.StringVar(&HTTPAddr, "http_addr", ":31234", "http listen address.")
	flag.StringVar(&GRPCAddr, "grpc_addr", ":31233", "grpc listen address.")
}

func main() {
	flag.Parse()

	httpSrv := server.NewHTTPServer(HTTPAddr)
	grpcSrv := server.NewGRPCServer(GRPCAddr)
	serverList := []transport.Server{httpSrv, grpcSrv}

	app := app.New(Name,
		&log.Conf{
			App:   Name,
			Level: "debug",
			Dev:   true,
		},
		serverList...,
	)

	{ // User service
		GreeterSrv := service.NewGreeterService()
		helloworld.RegisterGreeterHTTPServer(httpSrv.Container, GreeterSrv)
		helloworld.RegisterGreeterServer(grpcSrv.GetServe(), GreeterSrv)

		OpenapiSrv := service.NewOpenapiService()
		openapi.RegisterOpenapiHTTPServer(httpSrv.Container, OpenapiSrv)
		openapi.RegisterOpenapiServer(grpcSrv.GetServe(), OpenapiSrv)

		RulesSrv := service.NewRulesService()
		Rules_v1.RegisterRulesHTTPServer(httpSrv.Container, RulesSrv)
		Rules_v1.RegisterRulesServer(grpcSrv.GetServe(), RulesSrv)
	}

	if err := app.Run(context.TODO()); err != nil {
		panic(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop

	if err := app.Stop(context.TODO()); err != nil {
		panic(err)
	}
}
