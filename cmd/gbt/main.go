package main

import (
	"fmt"
	"log"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"

	"github.com/golang-base-template/cmd"
	gbthttp "github.com/golang-base-template/pkg/http"
	redisClient "github.com/golang-base-template/util/cache/client"
	"github.com/golang-base-template/util/config"
	databaseClient "github.com/golang-base-template/util/database/client"
	gbtserve "github.com/golang-base-template/util/serve"
)

func main() {
	err := config.InitConfig()
	if err != nil {
		msg := fmt.Sprintf("error when init http config: %+v", err)
		log.Fatalln(msg)
	}

	cfg := config.Get()
	// TODO: append the criticalDB with "gbt" to connect to db "gbt" for example
	criticalDB := []string{}
	nonCriticalDB := []string{}
	err = cmd.InitApp(cfg,
		databaseClient.DatabaseList{
			CriticalDatabase:    criticalDB,
			NonCriticalDatabase: nonCriticalDB,
		},
		redisClient.RedisList{
			CriticalRedis:    []string{},
			NonCriticalRedis: []string{},
		})
	if err != nil {
		msg := fmt.Sprintf("error when init http app: %+v", err)
		log.Fatalln(msg)
	}

	router := httprouter.New()
	gbthttp.Init()
	gbthttp.AssignRoutes(router)

	n := negroni.New()
	n.UseHandler(router)

	err = gbtserve.Serve(fmt.Sprintf(":%s", cfg.Port.GBT), n)
	if err != nil {
		log.Println("error when serve http app")
		return
	}
}
