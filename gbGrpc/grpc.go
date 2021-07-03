package gbGrpc

import (
	"context"
	"time"

	"github.com/gobench-io/gobench/executor"
	"github.com/gobench-io/gobench/executor/metrics"
	"google.golang.org/grpc"
)

type GbClientConn struct {
	grpc.ClientConn
	methodGraphsMap map[string][]metrics.Graph
	target          string
}

func Dial(target string, opts ...grpc.DialOption) (*GbClientConn, error) {
	return DialContext(context.Background(), target, opts...)
}

func DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (conn *GbClientConn, err error) {
	conn = &GbClientConn{
		methodGraphsMap: make(map[string][]metrics.Graph),
		target:          target,
	}

	c, err := grpc.DialContext(ctx, target, opts...)
	conn.ClientConn = *c

	return
}

func (cc *GbClientConn) setupMethod(method string) ([]metrics.Graph, error) {
	if graphs, ok := cc.methodGraphsMap[method]; ok {
		return graphs, nil
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

	cc.methodGraphsMap[method] = graphs

	group := metrics.Group{
		Name:   "gRPC (" + cc.target + ")",
		Graphs: graphs,
	}

	groups := []metrics.Group{
		group,
	}

	err := executor.Setup(groups)

	return graphs, err
}

func (cc *GbClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	graphs, _ := cc.setupMethod(method)

	begin := time.Now()

	err := cc.ClientConn.Invoke(ctx, method, args, reply, opts...)

	diff := time.Since(begin)

	latencyTitle := graphs[1].Metrics[0].Title

	countTitle := graphs[0].Metrics[0].Title
	if err != nil {
		countTitle = graphs[0].Metrics[1].Title
	}

	executor.Notify(latencyTitle, diff.Microseconds())
	executor.Notify(countTitle, 1)

	return err
}

type GbClientStream struct {
	grpc.ClientStream
	methodGraphsMap map[string][]metrics.Graph
	target          string
	method          string
}

func (cs *GbClientStream) setupMethod(target string, method string) (
	[]metrics.Graph, error,
) {
	if graphs, ok := cs.methodGraphsMap[method]; ok {
		return graphs, nil
	}

	graphs := []metrics.Graph{
		{
			Title: "New Stream",
			Unit:  "N",
			Metrics: []metrics.Metric{
				{
					Title: method + ".new_stream_ok", // success
					Type:  metrics.Counter,
				},
				{
					Title: method + ".new_stream_fail", // fail
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

	cs.methodGraphsMap[method] = graphs

	group := metrics.Group{
		Name:   "gRPC stream (" + cs.target + ")",
		Graphs: graphs,
	}

	groups := []metrics.Group{
		group,
	}

	err := executor.Setup(groups)

	return graphs, err
}

/**
what to record
CloseSend
SendMsg
RecvMsg
*/
func (cc *GbClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (
	grpc.ClientStream, error,
) {
	gcn := &GbClientStream{
		methodGraphsMap: make(map[string][]metrics.Graph),
		target:          cc.target,
		method:          method,
	}

	graphs, _ := gcn.setupMethod(cc.target, method)

	begin := time.Now()

	cs, err := cc.ClientConn.NewStream(ctx, desc, method, opts...)
	if err != nil {
		return cs, err
	}

	diff := time.Since(begin)

	gcn.ClientStream = cs

	return cs, err
}
