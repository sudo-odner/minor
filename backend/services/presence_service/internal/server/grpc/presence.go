package grpc

import (
	"context"

	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
)

func (s *ServerAPI) SetStatus(
	ctx context.Context,
	req *presencev1.SetStatusRequest,
) (*presencev1.SetStatusResponse, error) {
	const op = "server.grpc.SetStatus"
	// TODO: implement me
	return &presencev1.SetStatusResponse{
		Success: true,
	}, nil
}

func (s *ServerAPI) GetUserStatuses(
	ctx context.Context,
	req *presencev1.GetUserStatusesRequest,
) (*presencev1.GetUserStatusesResponse, error) {
	const op = "server.grpc.GetUserStatues"
	// TODO: implement me
	return &presencev1.GetUserStatusesResponse{
		Statuses: nil,
	}, nil
}
