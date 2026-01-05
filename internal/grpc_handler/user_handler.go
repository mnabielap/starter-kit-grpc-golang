package grpc_handler

import (
	"context"
	"strconv"

	pb "starter-kit-grpc-golang/api/gen/v1"
	"starter-kit-grpc-golang/internal/interceptor"
	"starter-kit-grpc-golang/internal/models"
	"starter-kit-grpc-golang/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	service service.UserService
}

func NewUserHandler(s service.UserService) *UserHandler {
	return &UserHandler{service: s}
}

// Helper to convert Model -> Proto
func convertUserToProto(u *models.User) *pb.UserResponse {
	return &pb.UserResponse{
		Id:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Role:            u.Role,
		IsEmailVerified: u.IsEmailVerified,
		CreatedAt:       timestamppb.New(u.CreatedAt),
		UpdatedAt:       timestamppb.New(u.UpdatedAt),
	}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	// RBAC: Admin Only
	if err := interceptor.AuthorizeAdmin(ctx); err != nil {
		return nil, err
	}

	user, err := h.service.CreateUser(req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// SET 201 CREATED
	grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(201)))

	return convertUserToProto(user), nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	// RBAC: Admin OR Self
	if err := interceptor.AuthorizeAdminOrSelf(ctx, req.Id); err != nil {
		return nil, err
	}

	user, err := h.service.GetUserByID(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return convertUserToProto(user), nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// RBAC: Admin Only
	if err := interceptor.AuthorizeAdmin(ctx); err != nil {
		return nil, err
	}

	filters := map[string]interface{}{
		"search": req.Search,
		"role":   req.Role,
		"scope":  req.Scope,
	}

	users, total, err := h.service.GetUsers(filters, req.Page, req.Limit, req.Sort)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var protoUsers []*pb.UserResponse
	for _, u := range users {
		protoUsers = append(protoUsers, convertUserToProto(&u))
	}

	// Calculate Total Pages
	limit := req.Limit
	if limit < 1 {
		limit = 10
	}
	totalPages := int32((total + int64(limit) - 1) / int64(limit))

	return &pb.ListUsersResponse{
		Results:      protoUsers,
		Page:         req.Page,
		Limit:        req.Limit,
		TotalPages:   totalPages,
		TotalResults: total,
	}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	// RBAC: Admin Only (Strict Mode)
	if err := interceptor.AuthorizeAdmin(ctx); err != nil {
		return nil, err
	}

	dto := service.UpdateUserDTO{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	user, err := h.service.UpdateUser(req.Id, dto)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertUserToProto(user), nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	// RBAC: Admin Only
	if err := interceptor.AuthorizeAdmin(ctx); err != nil {
		return nil, err
	}

	if err := h.service.DeleteUser(req.Id); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// SET 204 NO CONTENT
	grpc.SetHeader(ctx, metadata.Pairs("x-http-code", strconv.Itoa(204)))

	return &pb.DeleteUserResponse{Success: true}, nil
}