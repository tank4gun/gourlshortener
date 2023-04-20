package handlers

import (
	"context"
	"github.com/tank4gun/gourlshortener/internal/app/storage"
	"github.com/tank4gun/gourlshortener/internal/app/varprs"
	pb "github.com/tank4gun/gourlshortener/internal/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"net/http"
)

type URLShortenderServer struct {
	pb.UnimplementedURLShortenderServer
	// storage - storage.IRepository implementation
	storage storage.IRepository
	// baseURL - base URL for shorten URLs, i.e. http://localhost:8080
	baseURL string
	// deleteChannel - channel for RequestToDelete object to process
	deleteChannel chan RequestToDelete
}

func NewURLShortenderServer(storage storage.IRepository, baseURL string, deleteChannel chan RequestToDelete) *URLShortenderServer {
	return &URLShortenderServer{storage: storage, baseURL: baseURL, deleteChannel: deleteChannel}
}

func (s *URLShortenderServer) CreateShortUrl(ctx context.Context, in *pb.UrlToShortenRequest) (*pb.ShortenUrlResponse, error) {
	var response pb.ShortenUrlResponse
	shortURL, errorMessage, errorCode := s.storage.CreateShortURLByURL(in.Url, ctx.Value(UserIDCtxName).(uint))
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

func (s *URLShortenderServer) GetUrlById(ctx context.Context, in *pb.UrlByIdRequest) (*pb.UrlByIdResponse, error) {
	var response pb.UrlByIdResponse
	shortURL := in.ShortUrl
	id := ConvertShortURLToID(shortURL)
	url, errCode := s.storage.GetValueByKeyAndUserID(id, ctx.Value(UserIDCtxName).(uint))
	if errCode != 0 {
		return &response, status.Errorf(codes.NotFound, "Couldn't find url for id %s", shortURL)
	}
	response.OriginalUrl = url
	return &response, nil
}

func (s *URLShortenderServer) CreateShortenURLBatch(ctx context.Context, in *pb.BatchUrlRequest) (*pb.BatchUrlResponse, error) {
	var response pb.BatchUrlResponse

	var batchRequest []storage.BatchURLRequest
	for _, URL := range batchRequest {
		batchRequest = append(batchRequest, storage.BatchURLRequest{CorrelationID: URL.CorrelationID, OriginalURL: URL.OriginalURL})
	}
	resultURLs, errorMessage, errorCode := s.storage.CreateShortURLBatch(batchRequest, ctx.Value(UserIDCtxName).(uint), s.baseURL)

	if errorCode != 0 {
		return &response, status.Error(codes.Internal, errorMessage)
	}
	for _, resultURL := range resultURLs {
		response.Response = append(response.Response, &pb.CorrelationUrlResponse{CorrelationId: resultURL.CorrelationID, ShortUrl: resultURL.ShortURL})
	}
	return &response, nil
}

func (s *URLShortenderServer) GetAllUrls(ctx context.Context, in *emptypb.Empty) (*pb.FullInfoUrlBatchResponse, error) {
	var response pb.FullInfoUrlBatchResponse
	userID := ctx.Value(UserIDCtxName).(uint)
	responseList, errCode := s.storage.GetAllURLsByUserID(userID, s.baseURL)
	if errCode != http.StatusOK {
		return &response, status.Error(codes.Internal, "Got error while getting all URLs for user")
	}
	for _, responseItem := range responseList {
		response.Response = append(response.Response, &pb.FullInfoUrlBatchResponse_FullInfoUrl{ShortUrl: responseItem.ShortURL, OriginalUrl: responseItem.OriginalURL})
	}
	return &response, nil
}

func (s *URLShortenderServer) DeleteUrls(ctx context.Context, in *pb.DeleteUrlsRequest) (emptypb.Empty, error) {
	userID := ctx.Value(UserIDCtxName).(uint)
	URLsToDelete := make([]string, 0)
	for _, URL := range in.UrlsToDelete {
		URLsToDelete = append(URLsToDelete, URL.ShortUrl)
	}
	go func() {
		s.deleteChannel <- RequestToDelete{URLs: URLsToDelete, UserID: userID}
	}()
	return emptypb.Empty{}, nil
}

func (s *URLShortenderServer) Ping(ctx context.Context, in *emptypb.Empty) error {
	err := s.storage.Ping()
	if err != nil {
		return status.Error(codes.Unavailable, "Could not ping database")
	}
	return nil
}

func (s *URLShortenderServer) GetStats(ctx context.Context, in *emptypb.Empty) (*pb.StatsResponse, error) {
	var response pb.StatsResponse

	ipStr := ctx.Value("X-Real-IP").(string)
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
