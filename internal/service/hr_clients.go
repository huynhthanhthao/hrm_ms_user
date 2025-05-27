package service

import (
	hrPb "github.com/longgggwwww/hrm-ms-hr/ent/proto/entpb"

	"google.golang.org/grpc"
)

type HRServiceClients struct {
	Organization hrPb.OrganizationServiceClient
	HrExt        hrPb.ExtServiceClient
	Conn         *grpc.ClientConn
}

func (c *HRServiceClients) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}
