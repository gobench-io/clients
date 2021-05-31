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
	grpc.ClientConnInterface
	graphsMap map[string][]metrics.Graph
	target    string
}

func Dial(target string, opts ...grpc.DialOption) (*GbClientConn, error) {
	return DialContext(context.Background(), target, opts...)
}

func DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (conn *GbClientConn, err error) {
	conn = &GbClientConn{
		graphsMap: make(map[string][]metrics.Graph),
		target:    target,
	}

	conn.ClientConnInterface, err = grpc.DialContext(ctx, target, opts...)

	return
}

func (cc *GbClientConn) setupMethod(method string) ([]metrics.Graph, error) {
	if graphs, ok := cc.graphsMap[method]; ok {
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

	cc.graphsMap[method] = graphs

	group := metrics.Group{
		Name:   "gRPC (" + cc.target + ")",
		Graphs: graphs,
	}

	groups := []metrics.Group{
		group,
	}

	// waiting?
	err := executor.Setup(groups)

	return graphs, err
}

func (cc *GbClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	graphs, _ := cc.setupMethod(method)

	begin := time.Now()

	err := cc.ClientConnInterface.Invoke(ctx, method, args, reply, opts...)

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

func (cc *GbClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	begin := time.Now()

	log.Println("New stream")
	cs, err := cc.ClientConnInterface.NewStream(ctx, desc, method, opts...)

	// todo: record the duration
	_ = time.Since(begin)

	return cs, err
}
