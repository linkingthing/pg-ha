package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/zdnscloud/cement/rpc"

	"github.com/linkingthing/pg-ha/pkg/rpcserver"
)

const (
	RpcTimeout = 60 * time.Second
	RpcPort    = ":4400"
)

var (
	cmd    string
	master string
	slave  string
)

func init() {
	flag.StringVar(&cmd, "c", "start_ha", "cmd: start_ha/master_down/master_up")
	flag.StringVar(&master, "m", "", "kea_master config file dir")
	flag.StringVar(&slave, "s", "", "kea_master listen adress")
}

func startHA() error {
	if master == "" || slave == "" {
		return fmt.Errorf("master %s and slave %s must not be empty", master, slave)
	}

	_, err := rpc.RpcClient(master+RpcPort, RpcTimeout)
	if err != nil {
		return fmt.Errorf("connect to master failed: %s", err.Error())
	}

	slaveCli, err := rpc.RpcClient(slave+RpcPort, RpcTimeout)
	if err != nil {
		return fmt.Errorf("connect to slave failed: %s", err.Error())
	}

	return slaveCli.RpcCall("RPCServer.OnEvent", rpcserver.PGHACmdStartHA, nil)
}

func masterDown() error {
	if slave == "" {
		return fmt.Errorf("slave %s must not be empty", slave)
	}

	slaveCli, err := rpc.RpcClient(slave+RpcPort, RpcTimeout)
	if err != nil {
		return fmt.Errorf("connect to slave failed: %s", err.Error())
	}

	return slaveCli.RpcCall("RPCServer.OnEvent", rpcserver.PGHACmdMasterDown, nil)
}

func masterUp() error {
	if master == "" || slave == "" {
		return fmt.Errorf("master %s and slave %s must not be empty", master, slave)
	}

	_, err := rpc.RpcClient(master+RpcPort, RpcTimeout)
	if err != nil {
		return fmt.Errorf("connect to master failed: %s", err.Error())
	}

	slaveCli, err := rpc.RpcClient(slave+RpcPort, RpcTimeout)
	if err != nil {
		return fmt.Errorf("connect to slave failed: %s", err.Error())
	}

	return slaveCli.RpcCall("RPCServer.OnEvent", rpcserver.PGHACmdMasterUp, nil)
}

func main() {
	flag.Parse()

	var err error
	switch rpcserver.PGHACmd(cmd) {
	case rpcserver.PGHACmdStartHA:
		err = startHA()
	case rpcserver.PGHACmdMasterDown:
		err = masterDown()
	case rpcserver.PGHACmdMasterUp:
		err = masterUp()
	default:
		err = fmt.Errorf("unknown cmd %s", cmd)
	}

	if err != nil {
		fmt.Printf("run cmd %s failed: %s", cmd, err.Error())
	} else {
		fmt.Printf("cmd %s has been sent\n", cmd)
	}
}
