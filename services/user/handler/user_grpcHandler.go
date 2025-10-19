package handler

import (
	"context"
	"errors"

	"github.com/terkoizmy/golearn/pkg/pb/user"
	"github.com/terkoizmy/golearn/services/user/repository"
	"github.com/terkoizmy/golearn/services/user/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCHandler struct {
	user.UnimplementedUserServiceServer
	service service.UserService
}

func NewGRPCHandler(service service.UserService) *GRPCHandler {
	return &GRPCHandler{service: service}
}

func (h *GRPCHandler) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	u, err := h.service.GetUserByID(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &user.GetUserResponse{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (h *GRPCHandler) ValidateToken(ctx context.Context, req *user.ValidateTokenRequest) (*user.ValidateTokenResponse, error) {
	if req.Token == "" {
		return &user.ValidateTokenResponse{
			Valid:   false,
			Message: "token is required",
		}, nil
	}

	userID, err := h.service.ValidateToken(ctx, req.Token)
	if err != nil {
		return &user.ValidateTokenResponse{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}

	return &user.ValidateTokenResponse{
		Valid:   true,
		UserId:  userID,
		Message: "token is valid",
	}, nil
}
