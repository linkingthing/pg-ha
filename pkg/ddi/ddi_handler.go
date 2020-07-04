package ddi

import (
	"context"

	"google.golang.org/grpc"

	pb "github.com/linkingthing/pg-ha/pkg/proto"
)

type DDIHandler struct {
	grpcClient pb.DDICtrlManagerClient
	masterIP   string
}

func NewDDIHandler(ip string, conn *grpc.ClientConn) *DDIHandler {
	h := &DDIHandler{
		masterIP: ip,
	}
	if conn.Target() != "" {
		h.grpcClient = pb.NewDDICtrlManagerClient(conn)
	}
	return h
}

func (h *DDIHandler) MasterUp() error {
	if h.grpcClient != nil {
		_, err := h.grpcClient.MasterUp(context.TODO(), &pb.MasterUpRequest{MasterIp: h.masterIP})
		return err
	}

	return nil
}

func (h *DDIHandler) MasterDown() error {
	if h.grpcClient != nil {
		_, err := h.grpcClient.MasterDown(context.TODO(), &pb.MasterDownRequest{})
		return err
	}

	return nil
}
