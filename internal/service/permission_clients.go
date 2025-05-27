package service

import (
	permPb "github.com/huynhthanhthao/hrm_user_service/proto/permission"
	"google.golang.org/grpc"
)

type PermissionServiceClients struct {
	Conn     *grpc.ClientConn
	UserRole permPb.UserRoleServiceClient
	UserPerm permPb.UserPermServiceClient
	PermExt  permPb.ExtServiceClient
}

func (c *PermissionServiceClients) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}
