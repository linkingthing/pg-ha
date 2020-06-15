package handler

import (
	"time"

	_ "github.com/lib/pq"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/cement/rpc"
	"google.golang.org/grpc"

	"github.com/linkingthing/pg-ha/config"
)

type DBState string

const (
	DBStateStarting DBState = "starting"
	DBStateSyncing  DBState = "syncing"
	DBStateRunning  DBState = "running"
)

type PGHandler struct {
	proxy *PGProxy
	state DBState
	role  config.DBRole
}

func NewPGHandler(conf *config.PGHAConfig, conn *grpc.ClientConn) *PGHandler {
	return &PGHandler{
		proxy: newPGProxy(conf, conn),
		role:  conf.Server.Role,
	}
}

func (h *PGHandler) GetDBState() DBState {
	return h.state
}

func (h *PGHandler) GetDBRole() config.DBRole {
	return h.role
}

func (h *PGHandler) GetAnotherIP() string {
	return h.proxy.getAnotherIP()
}

func (h *PGHandler) SyncData() error {
	return h.proxy.syncData()
}

func (h *PGHandler) StartMaster(directly bool) error {
	log.Infof("start as master")
	h.state = DBStateStarting
	if !directly {
		if err := h.proxy.runDB(); err != nil {
			return err
		}

		if err := h.proxy.genPGConfigFile(false, false); err != nil {
			return err
		}

		if err := h.proxy.runDB(); err != nil {
			return err
		}

		h.state = DBStateSyncing
		log.Infof("sync data from master to slave")
		if err := h.proxy.syncData(); err != nil {
			return err
		}
	}

	if err := h.proxy.genPGConfigFile(true, false); err != nil {
		return err
	}

	if err := h.proxy.runDB(); err != nil {
		return err
	}

	h.state = DBStateRunning
	h.role = config.DBRoleMaster
	log.Infof("master run")
	return nil
}

func (h *PGHandler) StartSlave(directly bool) error {
	log.Infof("start as slave")
	h.state = DBStateStarting
	if err := h.proxy.stopDB(); err != nil {
		return err
	}

	if !directly {
		for {
			masterDBState, err := h.getMasterDBState()
			if err != nil {
				return err
			}

			if masterDBState == DBStateRunning {
				break
			} else {
				log.Infof("waiting for master to run, now its state is %v", masterDBState)
			}
			time.Sleep(time.Second)
		}
	}

	if err := h.proxy.genPGConfigFile(false, true); err != nil {
		return err
	}

	if err := h.proxy.runDB(); err != nil {
		return err
	}

	h.state = DBStateRunning
	h.role = config.DBRoleSlave
	log.Infof("slave run")
	return nil
}

func (h *PGHandler) StartSingle(role config.DBRole) error {
	log.Infof("start as %s", role)
	h.state = DBStateStarting
	if err := h.proxy.runDB(); err != nil {
		return err
	}

	if err := h.proxy.genPGConfigFile(false, false); err != nil {
		return err
	}

	if err := h.proxy.runDB(); err != nil {
		return err
	}

	h.state = DBStateRunning
	h.role = role
	log.Infof("%s run", role)
	return nil
}

func (h *PGHandler) getMasterDBState() (DBState, error) {
	var state DBState
	cli, err := rpc.RpcClient(h.proxy.getAnotherIP()+":4400", 3*time.Second)
	if err != nil {
		return state, err
	} else {
		defer cli.Close()
		if err := cli.RpcCall("RPCServer.GetDBState", "", &state); err != nil {
			return state, err
		}
	}
	return state, nil
}

func (h *PGHandler) StopDB() error {
	return h.proxy.stopDB()
}
