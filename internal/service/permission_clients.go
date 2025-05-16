package service

import (
	clientGrpc "github.com/huynhthanhthao/hrm_user_service/generated"
	"google.golang.org/grpc"
)

type PermissionServiceClients struct {
	Conn     *grpc.ClientConn
	UserRole clientGrpc.UserRoleServiceClient
	UserPerm clientGrpc.UserPermServiceClient
}

func (c *PermissionServiceClients) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}
