package grpc

// import (
// 	"context"
// 	"fmt"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// // HRClient encapsulates the gRPC client for HR service.
// type HRClient struct {
// 	client pb.HRServiceClient
// 	conn   *grpc.ClientConn
// }

// // NewHRClient creates a new HR gRPC client.
// func NewHRClient(addr string) (*HRClient, error) {
// 	conn, err := grpc.Dial(addr, grpc.WithInsecure()) // Use TLS in production
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to HR service: %w", err)
// 	}
// 	return &HRClient{
// 		client: pb.NewHRServiceClient(conn),
// 		conn:   conn,
// 	}, nil
// }

// // Close closes the gRPC connection.
// func (c *HRClient) Close() error {
// 	return c.conn.Close()
// }

// // ValidateCompany checks if a company_id exists in the HR service.
// func (c *HRClient) ValidateCompany(ctx context.Context, companyID int) error {
// 	resp, err := c.client.ValidateCompany(ctx, &pb.ValidateCompanyRequest{
// 		CompanyId: int32(companyID),
// 	})
// 	if err != nil {
// 		if status.Code(err) == codes.Unavailable {
// 			return fmt.Errorf("HR service unavailable: %w", err)
// 		}
// 		return fmt.Errorf("failed to validate company_id: %w", err)
// 	}
// 	if resp.Error != "" {
// 		return fmt.Errorf("HR service error: %s", resp.Error)
// 	}
// 	if !resp.Exists {
// 		return fmt.Errorf("company_id %d does not exist", companyID)
// 	}
// 	return nil
// }
