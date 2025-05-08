package grpcClient

type HRClient struct {
	// hrClient hrpb.ValidateCompanyServiceClient
	// conn     *grpc.ClientConn
}

func NewHRClient(addr string) (*HRClient, error) {
	// conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	return nil, err
	// }
	return &HRClient{
		// hrClient: hrpb.NewValidateCompanyServiceClient(conn),
		// conn:     conn,
	}, nil
}

// func (c *HRClient) ValidateCompany(ctx context.Context, companyID string) (bool, error) {
// resp, err := c.hrClient.ValidateCompany(ctx, &hrpb.ValidateCompanyRequest{CompanyId: companyID})
// if err != nil {
// 	return false, err
// }
// return resp.Exists, nil
// }

func (c *HRClient) Close() {
	// c.conn.Close()
}
