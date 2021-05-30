package gbGrpc

import (
	"context"
	"log"
	"time"

	"github.com/gobench-io/gobench/executor"
	"github.com/gobench-io/gobench/executor/metrics"
	"google.golang.org/grpc"
)

type GbClientConn struct {
	*grpc.ClientConn
	groups    []metrics.Group
	graphsMap map[string][]metrics.Graph
}

func Dial(target string, opts ...grpc.DialOption) (*GbClientConn, error) {
	return DialContext(context.Background(), target, opts...)
}

func DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (conn *GbClientConn, err error) {
	conn = &GbClientConn{}
	conn.ClientConn, err = grpc.DialContext(ctx, target, opts...)

	group := metrics.Group{
		Name: "gRPC (" + target + ")",
		Graphs: []metrics.Graph{
			{
				Title: "gRPC Response",
				Unit:  "N",
				Metrics: []metrics.Metric{
					{
						Title: "", // success
						Type:  metrics.Counter,
					},
					{
						Title: "", // fail
						Type:  metrics.Counter,
					},
				},
			},
			{
				Title: "Latency",
				Unit:  "Microsecond",
				Metrics: []metrics.Metric{
					{
						Title: "", // latency
						Type:  metrics.Histogram,
					},
				},
			},
		},
	}
	conn.groups = []metrics.Group{
		group,
	}

	return
}

func (cc *GbClientConn) setupMethod(method string) error {
	if _, ok := cc.graphsMap[method]; ok {
		return nil
	}

	graphs := []metrics.Graph{
		{
			Title: "gRPC Response",
			Unit:  "N",
			Metrics: []metrics.Metric{
				{
					Title: method + ".grpc_ok", // success
					Type:  metrics.Counter,
				},
				{
					Title: method + ".grpc_fail", // fail
					Type:  metrics.Counter,
				},
			},
		},
		{
			Title: "Latency",
			Unit:  "Microsecond",
			Metrics: []metrics.Metric{
				{
					Title: method + ".latency", // latency
					Type:  metrics.Histogram,
				},
			},
		},
	}

	cc.graphsMap[method] = graphs

	group := metrics.Group{
		Name:   "gRPC (" + cc.Target() + ")",
		Graphs: graphs,
	}

	groups := []metrics.Group{
		group,
	}

	err := executor.Setup(groups)

	return err
}

func (cc *GbClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	begin := time.Now()

	log.Println("Invoke")
	err := cc.ClientConn.Invoke(ctx, method, args, reply, opts...)

	// todo: record the duration
	_ = time.Since(begin)

	return err
}

func (cc *GbClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	begin := time.Now()

	log.Println("New stream")
	cs, err := cc.ClientConn.NewStream(ctx, desc, method, opts...)

	// todo: record the duration
	_ = time.Since(begin)

	return cs, err
}
