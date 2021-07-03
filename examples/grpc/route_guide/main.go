package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/gobench-io/clients/gbGrpc"
	"github.com/gobench-io/gobench/executor/scenario"
	"google.golang.org/grpc"

	pb "google.golang.org/grpc/examples/route_guide/routeguide"
)

func export() scenario.Vus {
	return scenario.Vus{
		{
			Nu:   5,
			Rate: 1000,
			Fu:   f,
		},
	}
}

func f(ctx context.Context, vui int) {
	address := "localhost:10000"

	// Set up a connection to the server.
	conn, err := gbGrpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewRouteGuideClient(conn)
	rect := &pb.Rectangle{
		Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000},
	}

	log.Printf("Looking for features within %v", rect)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.ListFeatures(ctx, rect)
	if err != nil {
		log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
	}

	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}
		log.Printf("Feature: name: %q, point:(%v, %v)", feature.GetName(),
			feature.GetLocation().GetLatitude(), feature.GetLocation().GetLongitude())
	}
}
