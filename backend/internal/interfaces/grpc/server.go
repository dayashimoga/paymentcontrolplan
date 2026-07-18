// Package grpc provides the gRPC server implementation for PCP.
// The gRPC server wraps the same application services as the HTTP API,
// providing high-performance internal communication.
//
// Proto generation (requires protoc + protoc-gen-go + protoc-gen-go-grpc):
//
//	protoc --go_out=. --go-grpc_out=. api/proto/v1/pcp.proto
package grpc

import (
	"context"
	"fmt"
	"net"

	appmch "github.com/paymentbridge/pcp/internal/application/merchant"
	apppay "github.com/paymentbridge/pcp/internal/application/payment"
	appprov "github.com/paymentbridge/pcp/internal/application/provider"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server wraps the gRPC server and application dependencies.
type Server struct {
	grpcServer      *grpc.Server
	merchantService *appmch.Service
	paymentService  *apppay.Service
	providerService *appprov.Service
	logger          *zap.Logger
	port            int
}

// NewServer creates and configures a new gRPC server.
func NewServer(
	merchantSvc *appmch.Service,
	paymentSvc *apppay.Service,
	providerSvc *appprov.Service,
	logger *zap.Logger,
	port int,
) *Server {
	s := &Server{
		merchantService: merchantSvc,
		paymentService:  paymentSvc,
		providerService: providerSvc,
		logger:          logger,
		port:            port,
	}

	// Create gRPC server with interceptors
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.loggingInterceptor),
	)

	// Register health service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(s.grpcServer, healthServer)
	healthServer.SetServingStatus("pcp", healthpb.HealthCheckResponse_SERVING)

	// Enable server reflection for debugging (disable in production)
	reflection.Register(s.grpcServer)

	// NOTE: Register generated service servers here after running protoc:
	// pcpv1.RegisterPaymentServiceServer(s.grpcServer, s)
	// pcpv1.RegisterProviderServiceServer(s.grpcServer, s)
	// pcpv1.RegisterMerchantServiceServer(s.grpcServer, s)

	return s
}

// Start begins listening for gRPC connections.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}
	s.logger.Info("gRPC server listening", zap.Int("port", s.port))
	return s.grpcServer.Serve(lis)
}

// Stop gracefully shuts down the gRPC server.
func (s *Server) Stop() {
	s.logger.Info("stopping gRPC server")
	s.grpcServer.GracefulStop()
}

// loggingInterceptor logs gRPC method calls.
func (s *Server) loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	s.logger.Info("gRPC call", zap.String("method", info.FullMethod))
	resp, err := handler(ctx, req)
	if err != nil {
		s.logger.Error("gRPC error", zap.String("method", info.FullMethod), zap.Error(err))
	}
	return resp, err
}
