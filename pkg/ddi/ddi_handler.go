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
	return &DDIHandler{
		grpcClient: pb.NewDDICtrlManagerClient(conn),
		masterIP:   ip,
	}
}

func (h *DDIHandler) MasterUp() error {
	_, err := h.grpcClient.MasterUp(context.TODO(), &pb.MasterUpRequest{MasterIp: h.masterIP})
	return err
}

func (h *DDIHandler) MasterDown() error {
	_, err := h.grpcClient.MasterDown(context.TODO(), &pb.MasterDownRequest{})
	return err
}
