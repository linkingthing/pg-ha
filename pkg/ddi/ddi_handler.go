package ddi

import (
	"context"

	"google.golang.org/grpc"

	"github.com/linkingthing/pg-ha/config"
	pb "github.com/linkingthing/pg-ha/pkg/proto"
)

type DDIHandler struct {
	grpcClient pb.DDICtrlManagerClient
	masterIP   string
	slaveIP    string
}

func NewDDIHandler(conf *config.PGHAConfig, ddiConn *grpc.ClientConn) *DDIHandler {
	h := &DDIHandler{
		masterIP: conf.Server.MasterIP,
		slaveIP:  conf.Server.SlaveIP,
	}
	if ddiConn.Target() != "" {
		h.grpcClient = pb.NewDDICtrlManagerClient(ddiConn)
	}
	return h
}

func (h *DDIHandler) MasterUp() error {
	if h.grpcClient != nil {
		if _, err := h.grpcClient.MasterUp(context.TODO(),
			&pb.DDICtrlRequest{MasterIp: h.masterIP, SlaveIp: h.slaveIP}); err != nil {
			return err
		}
	}

	return nil
}

func (h *DDIHandler) MasterDown() error {
	if h.grpcClient != nil {
		_, err := h.grpcClient.MasterDown(context.TODO(),
			&pb.DDICtrlRequest{MasterIp: h.masterIP, SlaveIp: h.slaveIP})
		return err
	}

	return nil
}
