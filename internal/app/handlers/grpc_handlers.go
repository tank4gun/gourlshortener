package handlers

import (
	"context"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
	pb "github.com/tank4gun/gourlshortener/internal/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"net/http"
	"strconv"
)

// ShortenderServer - grpc server struct
type ShortenderServer struct {
	pb.UnimplementedShortenderServer
	// storage - storage.IRepository implementation
	storage storage.IRepository
	// baseURL - base URL for shorten URLs, i.e. http://localhost:8080
	baseURL string
	// deleteChannel - channel for RequestToDelete object to process
	deleteChannel chan RequestToDelete
}

// NewShortenderServer - creates new grpc server instance
func NewShortenderServer(storage storage.IRepository, deleteChannel chan RequestToDelete) *ShortenderServer {
	return &ShortenderServer{storage: storage, baseURL: varprs.BaseURL, deleteChannel: deleteChannel}
}

// CreateShortURL - grpc handler, converts URL from request body to shorten one and saves into db
func (s *ShortenderServer) CreateShortURL(ctx context.Context, in *pb.UrlToShortenRequest) (*pb.ShortenUrlResponse, error) {
	var response pb.ShortenUrlResponse

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "Couldn't get data from context")
	}
	values := md.Get("UserID")
	if len(values) == 0 {
		return nil, status.Error(codes.Internal, "Couldn't get UserID from context")
	}
	userID, _ := strconv.Atoi(values[0])
	shortURL, errorMessage, errorCode := s.storage.CreateShortURLByURL(in.Url, uint(userID))
	if errorCode != 0 && errorCode != http.StatusConflict {
		return &response, status.Error(codes.InvalidArgument, errorMessage)
	}
	code := codes.OK
	if errorCode == http.StatusConflict {
		code = codes.AlreadyExists
	}
	response.ShortUrl = s.baseURL + shortURL
	return &response, status.Error(code, errorMessage)
}

// GetURLByID - grpc handler, returns full URL by its ID if it exists
func (s *ShortenderServer) GetURLByID(ctx context.Context, in *pb.UrlByIdRequest) (*pb.UrlByIdResponse, error) {
	var response pb.UrlByIdResponse
	shortURL := in.ShortUrl
	id := ConvertShortURLToID(shortURL)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "Couldn't get data from context")
	}
	values := md.Get("UserID")
	if len(values) == 0 {
		return nil, status.Error(codes.Internal, "Couldn't get UserID from context")
	}
	userID, _ := strconv.Atoi(values[0])
	url, errCode := s.storage.GetValueByKeyAndUserID(id, uint(userID))
	if errCode != 0 {
		return &response, status.Errorf(codes.NotFound, "Couldn't find url for id %s", shortURL)
	}
	response.OriginalUrl = url
	return &response, nil
}

// CreateShortenURLBatch - grpc handler, converts URL batch from json object to shorten one and saves into db
func (s *ShortenderServer) CreateShortenURLBatch(ctx context.Context, in *pb.BatchUrlRequest) (*pb.BatchUrlResponse, error) {
	var response pb.BatchUrlResponse

	var batchRequest []storage.BatchURLRequest
	for _, URL := range in.Request {
		batchRequest = append(batchRequest, storage.BatchURLRequest{CorrelationID: URL.CorrelationId, OriginalURL: URL.OriginalUrl})
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "Couldn't get data from context")
	}
	values := md.Get("UserID")
	if len(values) == 0 {
		return nil, status.Error(codes.Internal, "Couldn't get UserID from context")
	}
	userID, _ := strconv.Atoi(values[0])
	resultURLs, errorMessage, errorCode := s.storage.CreateShortURLBatch(batchRequest, uint(userID), s.baseURL)

	if errorCode != 0 {
		return &response, status.Error(codes.Internal, errorMessage)
	}
	for _, resultURL := range resultURLs {
		response.Response = append(response.Response, &pb.CorrelationUrlResponse{CorrelationId: resultURL.CorrelationID, ShortUrl: resultURL.ShortURL})
	}
	return &response, nil
}

// GetAllURLs - grpc handler, return all URLs for given User
func (s *ShortenderServer) GetAllURLs(ctx context.Context, in *emptypb.Empty) (*pb.FullInfoUrlBatchResponse, error) {
	var response pb.FullInfoUrlBatchResponse
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "Couldn't get data from context")
	}
	values := md.Get("UserID")
	if len(values) == 0 {
		return nil, status.Error(codes.Internal, "Couldn't get UserID from context")
	}
	userID, _ := strconv.Atoi(values[0])
	responseList, errCode := s.storage.GetAllURLsByUserID(uint(userID), s.baseURL)
	if errCode != http.StatusOK {
		return &response, status.Error(codes.Internal, "Got error while getting all URLs for user")
	}
	for _, responseItem := range responseList {
		response.Response = append(response.Response, &pb.FullInfoUrlBatchResponse_FullInfoUrl{ShortUrl: responseItem.ShortURL, OriginalUrl: responseItem.OriginalURL})
	}
	return &response, nil
}

// DeleteURLs - grpc handler, removes all URLs for given User
func (s *ShortenderServer) DeleteURLs(ctx context.Context, in *pb.DeleteUrlsRequest) (*emptypb.Empty, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "Couldn't get data from context")
	}
	values := md.Get("UserID")
	if len(values) == 0 {
		return nil, status.Error(codes.Internal, "Couldn't get UserID from context")
	}
	userID, _ := strconv.Atoi(values[0])
	URLsToDelete := make([]string, 0)
	for _, URL := range in.UrlsToDelete {
		URLsToDelete = append(URLsToDelete, URL.ShortUrl)
	}
	go func() {
		s.deleteChannel <- RequestToDelete{URLs: URLsToDelete, UserID: uint(userID)}
	}()
	return &emptypb.Empty{}, nil
}

// Ping - grpc handler, checks than connection to storage is alive
func (s *ShortenderServer) Ping(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.storage.Ping()
	if err != nil {
		return &emptypb.Empty{}, status.Error(codes.Unavailable, "Could not ping database")
	}
	return &emptypb.Empty{}, nil
}

// GetStats - grpc handler for statistics, return all URLs and Users number
func (s *ShortenderServer) GetStats(ctx context.Context, in *emptypb.Empty) (*pb.StatsResponse, error) {
	var response pb.StatsResponse
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "Couldn't get data from context")
	}
	values := md.Get("X-Real-IP")
	if len(values) == 0 {
		return nil, status.Error(codes.Internal, "Couldn't get X-Real-IP from context")
	}
	ipStr := values[0]
	requestIP := net.ParseIP(ipStr)
	if requestIP == nil {
		return nil, status.Error(codes.Aborted, "Got bad IP address")
	}
	_, ipNet, err := net.ParseCIDR(varprs.TrustedSubnet)
	if err != nil {
		return nil, status.Error(codes.Aborted, "Couldn't parse ipMask")
	}
	if !ipNet.Contains(requestIP) {
		return nil, status.Error(codes.Aborted, "Got bad IP address")
	}
	stats, errCode := s.storage.GetStats()
	if errCode != http.StatusOK {
		return nil, status.Error(codes.Internal, "Could not get stats")
	}
	response.Urls = int32(stats.URLs)
	response.Users = int32(stats.Users)
	return &response, nil
}
