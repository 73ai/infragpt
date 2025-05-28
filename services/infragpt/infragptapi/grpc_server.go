package infragptapi

import (
	"context"

	"github.com/priyanshujain/infragpt/services/infragpt"
	"github.com/priyanshujain/infragpt/services/infragpt/infragptapi/proto"
	"google.golang.org/grpc"
)

type grpcServer struct {
	proto.UnimplementedInfraGPTServiceServer
	svc infragpt.Service
}

func NewGRPCServer(svc infragpt.Service) *grpc.Server {
	server := grpc.NewServer()
	proto.RegisterInfraGPTServiceServer(server, &grpcServer{
		svc: svc,
	})
	return server
}

func (s *grpcServer) SendReply(ctx context.Context, req *proto.SendReplyCommand) (*proto.Status, error) {
	err := s.svc.SendReply(ctx, infragpt.SendReplyCommand{
		ConversationID: req.ConversationId,
		Message:        req.Message,
	})

	if err != nil {
		return &proto.Status{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.Status{
		Success: true,
		Error:   "",
	}, nil
}