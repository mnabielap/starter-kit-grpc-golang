package interceptor

import (
	"context"
	"net"
	"sync"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// IPRateLimiter holds rate limiters for each IP
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// RateLimitInterceptor creates the interceptor
func RateLimitInterceptor() grpc.UnaryServerInterceptor {
	// Allow 5 requests per second with burst of 20
	limiter := NewIPRateLimiter(5, 20)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		clientIP := getClientIP(ctx)
		
		if !limiter.GetLimiter(clientIP).Allow() {
			return nil, status.Error(codes.ResourceExhausted, "too many requests")
		}

		return handler(ctx, req)
	}
}

// getClientIP attempts to resolve IP from Metadata (Gateway) or Peer (Direct gRPC)
func getClientIP(ctx context.Context) string {
	// 1. Try X-Forwarded-For (From HTTP Gateway)
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		vals := md.Get("x-forwarded-for")
		if len(vals) > 0 {
			return vals[0]
		}
	}

	// 2. Try Peer Info
	if p, ok := peer.FromContext(ctx); ok {
		host, _, err := net.SplitHostPort(p.Addr.String())
		if err == nil {
			return host
		}
		return p.Addr.String()
	}

	return "unknown"
}