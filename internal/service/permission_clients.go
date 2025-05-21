package service

import (
	permissionPb "github.com/longgggwwww/hrm-ms-permission/ent/proto/entpb"
	"google.golang.org/grpc"
)

type PermissionServiceClients struct {
	Conn     *grpc.ClientConn
	UserRole permissionPb.UserRoleServiceClient
	UserPerm permissionPb.UserPermServiceClient
	PermExt  permissionPb.ExtServiceClient
}

func (c *PermissionServiceClients) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}
