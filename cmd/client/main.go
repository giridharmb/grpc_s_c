package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"google.golang.org/grpc"
	pb "grpc_s_c/pkg/grpc_s_c/pkg/pingpb" // adjust based on your generated package location
)

func main() {
	// Connect to the gRPC server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewPingServiceClient(conn)
	// In a real environment, determine your actual IP.
	vmIP := "192.168.1.100"

	// Start listening for messages from the server.
	go listenForMessages(client, vmIP)

	// Define an example JSON payload containing properties that could vary by VM type.
	examplePayload := map[string]interface{}{
		"vm_type":  "standard",
		"metadata": "example details",
	}
	payloadBytes, _ := json.Marshal(examplePayload)
	jsonPayload := string(payloadBytes)

	// Periodically send pings (every 5 seconds).
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		resp, err := client.SendPing(ctx, &pb.PingRequest{
			IpAddress:   vmIP,
			JsonPayload: jsonPayload,
		})
		cancel()
		if err != nil {
			log.Printf("error sending ping: %v", err)
		} else {
			log.Printf("Ping response: %s", resp.Status)
		}
	}
}

// listenForMessages opens a streaming RPC to receive pushed messages from the server.
func listenForMessages(client pb.PingServiceClient, ip string) {
	ctx := context.Background()
	stream, err := client.ListenForMessages(ctx, &pb.ListenRequest{IpAddress: ip})
	if err != nil {
		log.Fatalf("failed to open ListenForMessages stream: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("error receiving message: %v", err)
			return
		}
		log.Printf("Received message: %s", msg.JsonPayload)
	}
}
