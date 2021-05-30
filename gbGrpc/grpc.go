package gbGrpc

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

type GbClientConn struct {
	*grpc.ClientConn
	c int
}

func Dial(target string, opts ...grpc.DialOption) (*GbClientConn, error) {
	return DialContext(context.Background(), target, opts...)
}

func DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (conn *GbClientConn, err error) {
	conn = &GbClientConn{}
	conn.ClientConn, err = grpc.DialContext(ctx, target, opts...)

	return
}

func (cc *GbClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	log.Println("Invoke")
	err := cc.ClientConn.Invoke(ctx, method, args, reply, opts...)
	return err
}

func (cc *GbClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Println("New stream")
	cs, err := cc.ClientConn.NewStream(ctx, desc, method, opts...)

	return cs, err
}
