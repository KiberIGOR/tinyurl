package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/KiberIGOR/tinyurl/internal/auth"
	"github.com/KiberIGOR/tinyurl/internal/facade"
	grpcapi "github.com/KiberIGOR/tinyurl/internal/grpc"
	pb "github.com/KiberIGOR/tinyurl/internal/grpc/pb"
	"github.com/KiberIGOR/tinyurl/internal/repository"
	"github.com/KiberIGOR/tinyurl/internal/service"
	"github.com/stretchr/testify/require"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type memUserRepo struct {
	*repository.Memory
}

func TestGRPCShortenAndExpand(t *testing.T) {
	mem := repository.NewMemory()
	auth.Initialize(service.NewAuthService(memUserRepo{mem}, "test-secret"))

	svc := service.NewShortener(mem, "http://localhost/")
	fac := facade.NewShortener(svc, nil)

	srv := googlegrpc.NewServer(
		googlegrpc.UnaryInterceptor(auth.UnaryServerInterceptor),
	)
	pb.RegisterShortenerServiceServer(srv, grpcapi.NewShortenerServer(fac))

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(lis)
	t.Cleanup(func() { srv.GracefulStop() })

	conn, err := googlegrpc.NewClient(
		lis.Addr().String(),
		googlegrpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := pb.NewShortenerServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Создаём пользователя и получаем JWT.
	ctx = metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("authorization", ""),
	)

	var header metadata.MD

	short, err := client.ShortenURL(
		ctx,
		&pb.URLShortenRequest{
			Url: "https://example.com/page",
		},
		googlegrpc.Header(&header),
	)

	require.NoError(t, err)
	require.NotEmpty(t, short.GetResult())

	tokens := header.Get("authorization")
	require.NotEmpty(t, tokens)

	token := tokens[0]

	// Проверяем получение списка URL с новым JWT.
	authCtx := metadata.NewOutgoingContext(
		ctx,
		metadata.Pairs("authorization", token),
	)

	urls, err := client.ListUserURLs(
		authCtx,
		&emptypb.Empty{},
	)

	require.NoError(t, err)
	require.Len(t, urls.GetUrl(), 1)

	require.Equal(t, short.GetResult(), urls.GetUrl()[0].GetShortUrl())
	require.Equal(t, "https://example.com/page", urls.GetUrl()[0].GetOriginalUrl())

	require.NoError(t, err)
	require.NotEmpty(t, urls)

	// Проверяем ExpandURL.
	id := short.GetResult()[len("http://localhost/"):]

	got, err := client.ExpandURL(
		authCtx,
		&pb.URLExpandRequest{Id: id},
	)

	require.NoError(t, err)
	require.Equal(t, "https://example.com/page", got.GetResult())
}
