package ddi

import (
	"context"
	"time"

	"google.golang.org/grpc"

	"github.com/linkingthing/pg-ha/config"
	pb "github.com/linkingthing/pg-ha/pkg/proto"
	"github.com/zdnscloud/cement/log"
)

type DDIHandler struct {
	grpcMasterClient pb.DDICtrlManagerClient
	grpcSlaveClient  pb.DDICtrlManagerClient
	masterIP         string
	slaveIP          string
}

func NewDDIHandler(conf *config.PGHAConfig, masterConn *grpc.ClientConn, slaveConn *grpc.ClientConn) *DDIHandler {
	h := &DDIHandler{
		masterIP: conf.Server.MasterIP,
		slaveIP:  conf.Server.SlaveIP,
	}
	if masterConn.Target() != "" {
		h.grpcMasterClient = pb.NewDDICtrlManagerClient(masterConn)
	}
	if slaveConn.Target() != "" {
		h.grpcSlaveClient = pb.NewDDICtrlManagerClient(slaveConn)
	}

	return h
}

func (h *DDIHandler) MasterUp() error {
	if h.grpcSlaveClient != nil {
		if _, err := h.grpcSlaveClient.MasterUp(context.TODO(),
			&pb.DDICtrlRequest{MasterIp: h.masterIP, SlaveIp: h.slaveIP}); err != nil {
			return err
		}
	}

	if h.grpcMasterClient != nil {
		_, err := h.grpcMasterClient.MasterUp(context.TODO(),
			&pb.DDICtrlRequest{MasterIp: h.masterIP, SlaveIp: h.slaveIP})
		for err != nil {
			log.Infof("waiting for ddi_handler master response,now err:%s", err.Error())
			_, err = h.grpcMasterClient.MasterUp(context.TODO(),
				&pb.DDICtrlRequest{MasterIp: h.masterIP, SlaveIp: h.slaveIP})
			time.Sleep(time.Second)
		}
	}

	return nil
}

func (h *DDIHandler) MasterDown() error {
	if h.grpcSlaveClient != nil {
		_, err := h.grpcSlaveClient.MasterDown(context.TODO(), &pb.DDICtrlRequest{MasterIp: h.masterIP, SlaveIp: h.slaveIP})
		return err
	}

	return nil
}
