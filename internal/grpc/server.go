// Package grpc реализует gRPC API сервиса сокращения URL.
package grpc

import (
	"context"
	"errors"

	"github.com/KiberIGOR/tinyurl/internal/auth"
	"github.com/KiberIGOR/tinyurl/internal/facade"
	pb "github.com/KiberIGOR/tinyurl/internal/grpc/pb"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ShortenerServer обслуживает ShortenerService.
type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer
	facade *facade.Shortener
}

func NewShortenerServer(f *facade.Shortener) *ShortenerServer {
	return &ShortenerServer{facade: f}
}

func (s *ShortenerServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	if req == nil || req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	shortURL, err := s.facade.ShortenURL(ctx, userID, req.GetUrl())
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.URLShortenResponse{Result: shortURL}, nil
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	userID, _ := ctx.Value(auth.KeyUserID).(int)
	original, err := s.facade.ExpandURL(ctx, userID, req.GetId())
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.URLExpandResponse{Result: original}, nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, err := auth.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.facade.ListUserURLs(ctx, userID)
	if err != nil {
		if errors.Is(err, service.ErrNoContent) {
			return &pb.UserURLsResponse{}, nil
		}
		return nil, mapError(err)
	}
	out := make([]*pb.URLData, 0, len(items))
	for _, item := range items {
		out = append(out, &pb.URLData{
			ShortUrl:    item.ShortURL,
			OriginalUrl: item.OriginalURL,
		})
	}
	return &pb.UserURLsResponse{Url: out}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, "URL not found")
	case errors.Is(err, service.ErrGone):
		return status.Error(codes.FailedPrecondition, "URL deleted")
	case errors.Is(err, service.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}