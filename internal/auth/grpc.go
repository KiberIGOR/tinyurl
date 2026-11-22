package auth

import (
	"context"
	"strings"

	pb "github.com/KiberIGOR/tinyurl/internal/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const metadataAuthorization = "authorization"

// UnaryServerInterceptor извлекает JWT из metadata authorization и кладёт user id в контекст.
func UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	var err error
	switch info.FullMethod {
	case pb.ShortenerService_ExpandURL_FullMethodName:
		ctx, _, err = attachUserFromMetadata(ctx, authModePublic)
	case pb.ShortenerService_ShortenURL_FullMethodName:
		ctx, _, err = attachUserFromMetadata(ctx, authModeIssueIfMissing)
	default:
		ctx, _, err = attachUserFromMetadata(ctx, authModeRequired)
	}
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

type authMode int

const (
	authModeRequired authMode = iota
	authModeIssueIfMissing
	authModePublic
)

func attachUserFromMetadata(ctx context.Context, mode authMode) (context.Context, int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md.Get(metadataAuthorization)) == 0 {
		return handleMissingAuth(ctx, mode)
	}
	token := strings.TrimSpace(md.Get(metadataAuthorization)[0])
	if token == "" {
		return handleMissingAuth(ctx, mode)
	}
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	id := auth.GetUserID(ctx, token)
	if id <= 0 {
		if mode == authModeIssueIfMissing {
			return ensureUser(ctx)
		}
		if mode == authModePublic {
			return ctx, 0, nil
		}
		return ctx, 0, status.Error(codes.Unauthenticated, "invalid token")
	}
	return context.WithValue(ctx, KeyUserID, id), id, nil
}

func handleMissingAuth(ctx context.Context, mode authMode) (context.Context, int, error) {
	switch mode {
	case authModePublic:
		return ctx, 0, nil
	case authModeIssueIfMissing:
		return ensureUser(ctx)
	default:
		return ctx, 0, status.Error(codes.Unauthenticated, "missing authorization")
	}
}

func ensureUser(ctx context.Context) (context.Context, int, error) {
	if auth == nil {
		return ctx, 0, status.Error(codes.Internal, "auth not initialized")
	}
	id := auth.GetLastID(ctx)
	token, err := auth.BuildJWTString(ctx, id)
	if err != nil {
		return ctx, 0, status.Error(codes.Internal, err.Error())
	}
	_ = grpc.SetHeader(ctx, metadata.Pairs(metadataAuthorization, token))
	ctx = context.WithValue(ctx, KeyUserID, id)
	return ctx, id, nil
}

// UserIDFromContext возвращает id пользователя из контекста gRPC/HTTP.
func UserIDFromContext(ctx context.Context) (int, error) {
	id, ok := ctx.Value(KeyUserID).(int)
	if !ok || id <= 0 {
		return 0, status.Error(codes.Unauthenticated, "unauthorized")
	}
	return id, nil
}