package service

import (
	clientGrpc "github.com/huynhthanhthao/hrm_user_service/generated"
	"google.golang.org/grpc"
)

type HRServiceClients struct {
	Organization clientGrpc.OrganizationServiceClient
	HrExt        clientGrpc.ExtServiceClient
	Conn         *grpc.ClientConn
}

func (c *HRServiceClients) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}
