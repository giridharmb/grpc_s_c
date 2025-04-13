package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	pb "grpc_s_c/pkg/grpc_s_c/pkg/pingpb" // adjust the import path if necessary
)

// Client represents a virtual machine that pings the server.
type Client struct {
	ID        uint   `gorm:"primaryKey"`
	IPAddress string `gorm:"uniqueIndex"`
	LastSeen  int64
}

// PendingMessage represents a message stored for an offline client.
type PendingMessage struct {
	ID          uint `gorm:"primaryKey"`
	IPAddress   string
	JSONPayload string
	Delivered   bool
	CreatedAt   time.Time
	DeliveredAt *time.Time
}

// server implements pb.PingServiceServer.
type server struct {
	pb.UnimplementedPingServiceServer
	db          *gorm.DB
	redisClient *redis.Client
}

func main() {
	// ----- Set Up PostgreSQL (via GORM) -----
	dsn := "host=localhost user=myuser password=mypassword dbname=mydatabase port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	// AutoMigrate creates/updates tables.
	db.AutoMigrate(&Client{}, &PendingMessage{})

	// ----- Set Up Redis Client -----
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "my_redis_password",
		// In production, use appropriate options (authentication, timeouts, etc.)
	})

	// Create the server instance with both DB and Redis.
	s := &server{
		db:          db,
		redisClient: redisClient,
	}

	// ----- Start Background Tasks -----
	go s.cleanupClients()         // remove clients inactive for > 5 seconds from DB
	go s.cleanupPendingMessages() // delete pending messages older than 1 hour
	go s.eventHubListener()       // simulate an external event feed

	// ----- Start gRPC Server -----
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterPingServiceServer(grpcServer, s)
	log.Println("gRPC server running on :50051")
	grpcServer.Serve(lis)
}

// SendPing handles the pings from the client. It upserts the client record,
// updates Redis with a short TTL for the client's active status,
// and delivers any pending messages via Redis Pub/Sub.
func (s *server) SendPing(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	now := time.Now().Unix()

	// Upsert (insert or update) the client record.
	var client Client
	if err := s.db.Where("ip_address = ?", req.IpAddress).First(&client).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			client = Client{
				IPAddress: req.IpAddress,
				LastSeen:  now,
			}
			s.db.Create(&client)
		} else {
			log.Printf("DB error: %v", err)
			return &pb.PingResponse{Status: "error"}, err
		}
	} else {
		client.LastSeen = now
		s.db.Save(&client)
	}

	// Update Redis: set key "active:<ip>" with a TTL of 5 seconds.
	redisKey := "active:" + req.IpAddress
	if err := s.redisClient.Set(ctx, redisKey, now, 5*time.Second).Err(); err != nil {
		log.Printf("failed to update redis key %s: %v", redisKey, err)
	}

	// Check for any pending messages in the database.
	var messages []PendingMessage
	s.db.Where("ip_address = ? AND delivered = false", req.IpAddress).Find(&messages)
	// If there are pending messages, publish them on the Redis channel for this client.
	redisChannel := "client:" + req.IpAddress + ":messages"
	for _, m := range messages {
		if err := s.redisClient.Publish(ctx, redisChannel, m.JSONPayload).Err(); err != nil {
			log.Printf("failed publishing pending message for %s: %v", req.IpAddress, err)
		} else {
			deliveredAt := time.Now()
			m.Delivered = true
			m.DeliveredAt = &deliveredAt
			s.db.Save(&m)
			log.Printf("Delivered pending message to %s: %s", req.IpAddress, m.JSONPayload)
		}
	}

	// Optionally, process the incoming JSON payload.
	var extraPayload map[string]interface{}
	if err := json.Unmarshal([]byte(req.JsonPayload), &extraPayload); err == nil {
		log.Printf("Received ping payload from %s: %+v", req.IpAddress, extraPayload)
	} else if req.JsonPayload != "" {
		log.Printf("Received non-JSON payload from %s: %s", req.IpAddress, req.JsonPayload)
	}

	return &pb.PingResponse{Status: "pong", JsonPayload: "{}"}, nil
}

// ListenForMessages sets up a Redis Pub/Sub subscription so that the server can push messages
// to the client. This is a server-streaming RPC.
func (s *server) ListenForMessages(req *pb.ListenRequest, stream pb.PingService_ListenForMessagesServer) error {
	ctx := context.Background()
	redisChannel := "client:" + req.IpAddress + ":messages"
	pubsub := s.redisClient.Subscribe(ctx, redisChannel)
	defer pubsub.Close()

	log.Printf("Client %s subscribed to messages on channel %s", req.IpAddress, redisChannel)
	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			// Each message received from Redis is sent to the client.
			outMsg := &pb.Message{JsonPayload: msg.Payload}
			if err := stream.Send(outMsg); err != nil {
				log.Printf("failed sending message to %s: %v", req.IpAddress, err)
			}
		case <-stream.Context().Done():
			log.Printf("Client %s unsubscribed from messages", req.IpAddress)
			return nil
		}
	}
}

// cleanupClients periodically removes database records for clients that haven't pinged in > 5 seconds.
func (s *server) cleanupClients() {
	for {
		threshold := time.Now().Unix() - 5
		s.db.Where("last_seen < ?", threshold).Delete(&Client{})
		time.Sleep(5 * time.Second)
	}
}

// cleanupPendingMessages removes pending messages older than 1 hour.
func (s *server) cleanupPendingMessages() {
	for {
		cutoff := time.Now().Add(-1 * time.Hour)
		s.db.Where("created_at < ? AND delivered = false", cutoff).Delete(&PendingMessage{})
		time.Sleep(10 * time.Minute)
	}
}

// eventHubListener simulates an external event feed (e.g. Azure Event Hub).
// It produces messages with a generic JSON payload.
func (s *server) eventHubListener() {
	ticker := time.NewTicker(10 * time.Second)
	ctx := context.Background()
	for range ticker.C {
		// Simulated external message.
		extMsg := externalMessage{
			IPAddress: "192.168.1.100", // Example IP; in a real system, this comes in the event payload.
			Data:      `{"update": "Important external update", "severity": "high"}`,
		}
		s.processExternalMessage(ctx, extMsg)
	}
}

// externalMessage is a simple struct representing an external event.
type externalMessage struct {
	IPAddress string
	Data      string // A JSON string payload.
}

// processExternalMessage looks up the target client by its IP using Redis.
// If the client is active, it publishes the message on the Redis channel so that
// the client (who is subscribed) receives it; otherwise, the message is stored in the DB.
func (s *server) processExternalMessage(ctx context.Context, msg externalMessage) {
	redisKey := "active:" + msg.IPAddress
	active, err := s.redisClient.Exists(ctx, redisKey).Result()
	if err != nil {
		log.Printf("redis exists check failed for %s: %v", msg.IPAddress, err)
		return
	}
	redisChannel := "client:" + msg.IPAddress + ":messages"
	if active > 0 {
		// Client is active, so publish the message on its channel.
		if err := s.redisClient.Publish(ctx, redisChannel, msg.Data).Err(); err != nil {
			log.Printf("failed publishing external message to %s: %v", msg.IPAddress, err)
		} else {
			log.Printf("Delivered external message to active client %s: %s", msg.IPAddress, msg.Data)
		}
	} else {
		// Client is inactive; store the message as pending.
		pending := PendingMessage{
			IPAddress:   msg.IPAddress,
			JSONPayload: msg.Data,
			Delivered:   false,
			CreatedAt:   time.Now(),
		}
		s.db.Create(&pending)
		log.Printf("Stored external message for inactive client %s", msg.IPAddress)
	}
}
