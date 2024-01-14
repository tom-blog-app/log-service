package service

import (
	"context"
	"fmt"
	logProto "github.com/tom-blog-app/blog-proto/log"
	"github.com/tom-blog-app/log-service/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"time"
)

var collection = os.Getenv("LOG_MONGO_DB")

type LogServer struct {
	logProto.UnimplementedLogServiceServer
	Client *mongo.Client
}

func (s *LogServer) CreateLog(ctx context.Context, req *logProto.LogRequest) (*logProto.LogResponse, error) {
	log := req.GetLog()
	logModel := models.Log{
		ID:        primitive.NewObjectID().Hex(),
		Name:      log.GetName(),
		Content:   log.GetContent(),
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}

	// Insert the post into the database
	collection := s.Client.Database(collection).Collection(collection)
	result, err := collection.InsertOne(ctx, logModel)
	if err != nil {
		return nil, fmt.Errorf("could not insert post: %v", err)
	}

	id := result.InsertedID.(primitive.ObjectID).Hex()

	// Return a successful response
	res := &logProto.LogResponse{
		Log: &logProto.Log{
			Id:        id,
			Name:      log.GetName(),
			Content:   log.GetContent(),
			CreatedAt: log.CreatedAt,
		},
	}
	return res, nil
}

func (s *LogServer) DeleteLog(ctx context.Context, req *logProto.GetLogRequest) (*logProto.LogDeleteResponse, error) {
	collection := s.Client.Database(collection).Collection(collection)
	_, err := collection.DeleteOne(ctx, bson.M{"_id": req.GetId()})
	if err != nil {
		return nil, fmt.Errorf("could not delete log: %v", err)
	}

	return &logProto.LogDeleteResponse{
		Id:      req.GetId(),
		Success: true,
	}, nil
}

func (s *LogServer) ListLog(ctx context.Context, req *logProto.GetLogListRequest) (*logProto.ListLogsResponse, error) {
	collection := s.Client.Database(collection).Collection(collection)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("could not list logs: %v", err)
	}
	defer cursor.Close(ctx)

	var logs []*logProto.Log
	for cursor.Next(ctx) {
		var logModel models.Log
		err := cursor.Decode(&logModel)
		if err != nil {
			return nil, fmt.Errorf("could not decode log: %v", err)
		}

		log := &logProto.Log{
			Id:        logModel.ID,
			Name:      logModel.Name,
			Content:   logModel.Content,
			CreatedAt: timestamppb.New(logModel.CreatedAt),
		}
		logs = append(logs, log)
	}

	return &logProto.ListLogsResponse{Logs: logs}, nil
}

func (s *LogServer) ListLogByDate(ctx context.Context, req *logProto.GetLogListRequestByDate) (*logProto.ListLogsResponse, error) {
	startDate := req.GetStartDate().AsTime()
	endDate := req.GetEndDate().AsTime()

	// Create a filter that matches logs where the CreatedAt field is greater than or equal to the start date
	// and less than or equal to the end date
	filter := bson.M{
		"created_at": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	collection := s.Client.Database(collection).Collection(collection)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("could not list logs: %v", err)
	}
	defer cursor.Close(ctx)

	var logs []*logProto.Log
	for cursor.Next(ctx) {
		var logModel models.Log
		err := cursor.Decode(&logModel)
		if err != nil {
			return nil, fmt.Errorf("could not decode log: %v", err)
		}

		log := &logProto.Log{
			Id:        logModel.ID,
			Name:      logModel.Name,
			Content:   logModel.Content,
			CreatedAt: timestamppb.New(logModel.CreatedAt),
		}
		logs = append(logs, log)
	}

	return &logProto.ListLogsResponse{Logs: logs}, nil
}
