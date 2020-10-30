package rpcserver

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/looplab/fsm"
	"github.com/zdnscloud/cement/log"
	"github.com/zdnscloud/cement/rpc"
	"google.golang.org/grpc"

	"github.com/linkingthing/pg-ha/config"
	"github.com/linkingthing/pg-ha/pkg/ddi"
	"github.com/linkingthing/pg-ha/pkg/pg"
)

type PGHACmd string

const (
	PGHACmdStartSingle PGHACmd = "start_single"
	PGHACmdStartHA     PGHACmd = "start_ha"
	PGHACmdMasterDown  PGHACmd = "master_down"
	PGHACmdMasterUp    PGHACmd = "master_up"
	PGHACmdQueryState  PGHACmd = "query_state"

	EventRunAsSingle     string = "run_as_single"
	EventRunAsMaster     string = "run_as_master"
	EventRunAsSlave      string = "run_as_slave"
	EventTurnToMaster    string = "turn_to_master"
	EventTurnToSlave     string = "turn_to_slave"
	EventTurnToTmpMaster string = "turn_to_tmp_master"
	EventTurnBackToSlave string = "turn_back_to_slave"

	RpcTimeout = 60 * time.Second
	RPCPort    = ":4400"
)

type RPCServer struct {
	pgHandler  *pg.PGHandler
	ddiHandler *ddi.DDIHandler
	recentCmd  PGHACmd
	fsm        *fsm.FSM
	eventChan  chan string
	fsmState   string
}

func Run(conf *config.PGHAConfig, agentConn *grpc.ClientConn, ddiMasterConn *grpc.ClientConn, ddiSlaveConn *grpc.ClientConn) error {
	s := &RPCServer{
		pgHandler:  pg.NewPGHandler(conf, agentConn),
		ddiHandler: ddi.NewDDIHandler(conf, ddiMasterConn, ddiSlaveConn),
		eventChan:  make(chan string, 10),
	}

	s.fsm = fsm.NewFSM(
		string(config.DBRoleSingle),
		fsm.Events{
			{Name: EventRunAsSingle, Src: []string{string(config.DBRoleSingle), string(config.DBRoleMaster),
				string(config.DBRoleSlave), string(config.DBRoleTmpMaster)}, Dst: string(config.DBRoleSingle)},
			{Name: EventRunAsMaster, Src: []string{string(config.DBRoleSingle), string(config.DBRoleMaster),
				string(config.DBRoleSlave), string(config.DBRoleTmpMaster)}, Dst: string(config.DBRoleMaster)},
			{Name: EventRunAsSlave, Src: []string{string(config.DBRoleSingle), string(config.DBRoleMaster),
				string(config.DBRoleSlave), string(config.DBRoleTmpMaster)}, Dst: string(config.DBRoleSlave)},
			{Name: EventTurnToMaster, Src: []string{string(config.DBRoleSingle), string(config.DBRoleMaster)},
				Dst: string(config.DBRoleMaster)},
			{Name: EventTurnToSlave, Src: []string{string(config.DBRoleSingle), string(config.DBRoleSlave),
				string(config.DBRoleTmpMaster)}, Dst: string(config.DBRoleSlave)},
			{Name: EventTurnToTmpMaster, Src: []string{string(config.DBRoleSlave)}, Dst: string(config.DBRoleTmpMaster)},
			{Name: EventTurnBackToSlave, Src: []string{string(config.DBRoleSingle), string(config.DBRoleTmpMaster)},
				Dst: string(config.DBRoleSlave)},
		},

		fsm.Callbacks{
			EventRunAsSingle:     func(e *fsm.Event) { s.runAsSingle(e) },
			EventRunAsMaster:     func(e *fsm.Event) { s.runAsMaster(e) },
			EventRunAsSlave:      func(e *fsm.Event) { s.runAsSlave(e) },
			EventTurnToMaster:    func(e *fsm.Event) { s.turnToMaster(e) },
			EventTurnToSlave:     func(e *fsm.Event) { s.turnToSlave(e) },
			EventTurnToTmpMaster: func(e *fsm.Event) { s.turnToTmpMaster(e) },
			EventTurnBackToSlave: func(e *fsm.Event) { s.turnBackToSlave(e) },
		},
	)

	if _, err := rpc.RunRpcServer(s, "0.0.0.0"+RPCPort); err != nil {
		return err
	}

	if err := s.init(); err != nil {
		return err
	}

	go s.run()
	return nil
}

func (s *RPCServer) init() error {
	isHA := false
	if s.pgHandler.GetDBRole() == config.DBRoleMaster {
		isHA = checkIfExist("\"postgres: walsender\"")
	} else {
		isHA = checkIfExist("\"postgres: walreceiver\"")
	}

	if err := s.pgHandler.StopDB(); err != nil {
		return err
	}

	states := s.queryDestState()
	if isHA == false && (len(states) == 0 || dbIsRunning(states) == false ||
		isRoleMatched(states, config.DBRoleSingle) || isRoleMatched(states, config.DBRoleTmpMaster)) {
		s.eventChan <- EventRunAsSingle
		return nil
	}

	s.recentCmd = PGHACmdStartHA
	if role := s.pgHandler.GetDBRole(); role == config.DBRoleMaster {
		s.eventChan <- EventTurnToMaster
	} else if role == config.DBRoleSlave {
		s.eventChan <- EventTurnToSlave
	}

	return nil
}

func checkIfExist(processName string) bool {
	out, _ := exec.Command("bash", "-c", "ps -ef | grep "+processName+" | grep -v grep").Output()
	return len(out) > 0
}

func (s *RPCServer) queryDestState() []string {
	anotherIP := s.pgHandler.GetAnotherIP()
	if anotherIP == "" {
		return nil
	}

	cli, err := rpc.RpcClient(anotherIP+RPCPort, RpcTimeout)
	if err != nil {
		return nil
	}

	var states []string
	if err := cli.RpcCall("RPCServer.OnEvent", PGHACmdQueryState, &states); err != nil {
		return nil
	}

	return states
}

func dbIsRunning(states []string) bool {
	return states[0] == string(pg.DBStateRunning)
}

func isRoleMatched(states []string, role config.DBRole) bool {
	return states[1] == string(role)
}

func (s *RPCServer) run() {
	for {
		e := <-s.eventChan
		log.Infof(fmt.Sprintf("get event = %s, current state = %s", e, s.fsm.Current()))
		if err := s.fsm.Event(e); err != nil {
			log.Errorf(err.Error())
		} else {
			log.Infof(fmt.Sprintf("turn to %s", s.fsm.Current()))
			s.setState(s.fsm.Current())
		}
	}
}

func (s *RPCServer) setState(state string) {
	s.fsmState = state
}

func (s *RPCServer) OnEvent(cmd string, states *[]string) error {
	var err error
	switch PGHACmd(cmd) {
	case PGHACmdStartHA, PGHACmdStartSingle, PGHACmdMasterDown, PGHACmdMasterUp:
		err = s.event(PGHACmd(cmd))
	case PGHACmdQueryState:
		err = s.queryState(states)
	default:
		err = fmt.Errorf("unsupported cmd: %s\n", cmd)
	}

	if err != nil {
		log.Errorf(fmt.Sprintf("handler event with cmd %s failed: %s", cmd, err.Error()))
	}

	return err
}

func (s *RPCServer) queryState(states *[]string) error {
	*states = append(*states, string(s.pgHandler.GetDBState()), string(s.pgHandler.GetDBRole()), string(s.recentCmd))
	return nil
}

func (s *RPCServer) event(cmd PGHACmd) error {
	log.Infof(fmt.Sprintf("get cmd = %s", cmd))
	e, err := s.pharseCmdToEvent(cmd)
	if err != nil {
		return err
	}

	select {
	case s.eventChan <- e:
		s.recentCmd = cmd
		return nil
	case <-time.After(time.Second * 5):
		return fmt.Errorf("send event %s time out", cmd)
	}
}

func (s *RPCServer) pharseCmdToEvent(cmd PGHACmd) (string, error) {
	role := config.Get().Server.Role
	switch cmd {
	case PGHACmdStartHA:
		if role == config.DBRoleMaster {
			return EventRunAsMaster, nil
		} else if role == config.DBRoleSlave {
			return EventRunAsSlave, nil
		}
	case PGHACmdStartSingle:
		return EventRunAsSingle, nil
	case PGHACmdMasterDown:
		if role == config.DBRoleSlave {
			return EventTurnToTmpMaster, nil
		}
	case PGHACmdMasterUp:
		if role == config.DBRoleMaster {
			return EventTurnToMaster, nil
		} else if role == config.DBRoleSlave {
			return EventTurnBackToSlave, nil
		}
	}
	return "", fmt.Errorf("parse cmd to event failed with unsupported cmd: %s", cmd)
}

func (s *RPCServer) runAsSingle(e *fsm.Event) {
	//TODO pg_resetxlog
	if err := s.pgHandler.StartSingle(config.DBRoleSingle); err != nil {
		log.Errorf("start single failed: %s", err.Error())
		panic(err)
	}
}

func (s *RPCServer) runAsMaster(e *fsm.Event) {
	if err := s.pgHandler.StartMaster(false); err != nil {
		log.Errorf("run as master failed: %s", err.Error())
		panic(err)
	}
}

func (s *RPCServer) runAsSlave(e *fsm.Event) {
	masterConn, err := rpc.RpcClient(s.pgHandler.GetAnotherIP()+RPCPort, RpcTimeout)
	if err != nil {
		log.Errorf(fmt.Sprintf("connect to master failed: %s", err.Error()))
		s.rollBackTo(s.fsmState)
		log.Warnf("roll back to %s", s.fsmState)
		return
	}

	if err := masterConn.RpcCall("RPCServer.OnEvent", PGHACmdStartHA, nil); err != nil {
		log.Errorf("start ha failed: %s", err.Error())
		s.rollBackTo(s.fsmState)
		log.Warnf("roll back to %s", s.fsmState)
		return
	}

	if err := s.pgHandler.StartSlave(false); err != nil {
		log.Errorf(err.Error())
		panic(err)
	}
}

func (s *RPCServer) turnToMaster(e *fsm.Event) {
	states := s.queryDestState()
	hot := false

	if states == nil || (isRoleMatched(states, config.DBRoleSlave) && dbIsRunning(states) && s.needHotStart(states)) {
		log.Infof("++hot start")
		hot = true
	} else {
		log.Infof("++cold start")
	}

	if err := s.pgHandler.StartMaster(hot); err != nil {
		log.Errorf("turn to master failed: %s", err.Error())
		panic(err)
	}
}

func (s *RPCServer) needHotStart(states []string) bool {
	return (s.recentCmd == PGHACmdStartHA || s.recentCmd == PGHACmdMasterUp) &&
		(states[2] == string(PGHACmdStartHA) || states[2] == string(PGHACmdMasterUp))
}

func (s *RPCServer) turnToSlave(e *fsm.Event) {
	if err := s.pgHandler.StartSlave(len(s.queryDestState()) == 0); err != nil {
		log.Errorf("start slave failed: %s", err.Error())
		panic(err)
	}
}

func (s *RPCServer) turnToTmpMaster(e *fsm.Event) {
	if err := s.pgHandler.StartSingle(config.DBRoleTmpMaster); err != nil {
		log.Errorf("start tmp master failed: %s", err.Error())
		panic(err)
	}

	if err := s.ddiHandler.MasterDown(); err != nil {
		log.Errorf("send master down to ddi failed: %s", err.Error())
		panic(err)
	}
}

func (s *RPCServer) turnBackToSlave(e *fsm.Event) {
	masterConn, err := rpc.RpcClient(s.pgHandler.GetAnotherIP()+RPCPort, RpcTimeout)
	if err != nil {
		log.Errorf(fmt.Sprintf("connect to master failed: %s", err.Error()))
		s.rollBackTo(s.fsmState)
		log.Warnf("roll back to %s", s.fsmState)
		return
	}

	if err := masterConn.RpcCall("RPCServer.ShutDown", "", nil); err != nil {
		log.Errorf(fmt.Sprintf("shutdwon master: %s", err.Error()))
	}

	if err := s.pgHandler.SyncData(); err != nil {
		log.Errorf("turn to slave when sync data failed: %s", err.Error())
		panic(err)
	}

	if err := masterConn.RpcCall("RPCServer.OnEvent", PGHACmdMasterUp, nil); err != nil {
		log.Errorf("turn to slave when master up failed: %s", err.Error())
		s.rollBackTo(string(config.DBRoleTmpMaster))
		log.Warnf("roll back to %s", config.DBRoleTmpMaster)
		return
	}

	if err := s.pgHandler.StartSlave(len(s.queryDestState()) == 0); err != nil {
		log.Errorf("start slave failed: %s", err.Error())
		return
	}

	if err := s.ddiHandler.MasterUp(); err != nil {
		log.Errorf("send master up to ddi failed: %s", err.Error())
		panic(err)
	}
}

func (s *RPCServer) rollBackTo(state string) {
	s.fsm.SetState(state)
}

func (s *RPCServer) ShutDown(unusedArg string, noreturn *string) error {
	return s.pgHandler.StopDB()
}

func (s *RPCServer) GetDBState(unusedArg string, state *pg.DBState) error {
	*state = s.pgHandler.GetDBState()
	return nil
}
