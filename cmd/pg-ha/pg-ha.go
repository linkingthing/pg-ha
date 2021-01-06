package main

import (
	"flag"

	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/cement/signal"
	"google.golang.org/grpc"

	"github.com/linkingthing/pg-ha/config"
	"github.com/linkingthing/pg-ha/pkg/rpcserver"
)

var (
	confFile string
)

func init() {
	flag.StringVar(&confFile, "c", "pg-ha.conf", "pg ha configure file")
}

func main() {
	flag.Parse()
	log.InitLogger(log.Debug)

	conf, err := config.Load(confFile)
	if err != nil {
		log.Fatalf("load config file failed: %s", err.Error())
	}

	agentConn, err := grpc.Dial(conf.PGAgent.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("conn pg agent failed: %s", err.Error())
	}

	defer agentConn.Close()

	if err := rpcserver.Run(conf, agentConn); err != nil {
		log.Fatalf("new rpc server failed: %s", err.Error())
	}

	signal.WaitForInterrupt(nil)
}
