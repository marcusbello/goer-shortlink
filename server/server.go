package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teris-io/shortid"
	"go.uber.org/zap"
	"goer-shortlink/data"
	"goer-shortlink/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"time"
)

// build options
var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	port     = flag.Int("port", 50051, "The server port")
)

// urlShortenerServer is used to implement UrlShortenerService
type urlShortenerServer struct {
	proto.UnimplementedUrlShortenerServer

	logger    *zap.SugaredLogger
	shortCode *shortid.Shortid
	db        *pgxpool.Pool
	//TODO: implement logger and tracing
}

// ShortLink controls our proto.ShortLink service it accepts input url and returns the short url
func (s *urlShortenerServer) ShortLink(ctx context.Context, in *proto.Request) (*proto.Response, error) {
	var resp proto.Response
	var message proto.Message
	shortCode := s.shortCode.MustGenerate() // shortId package handle the short code generation
	s.logger.Infof("generated -> %v -> %v", shortCode, in.Input)

	// TODO: validate input here
	if in.Input == "" {
		resp.Code = 400
		resp.Error = "empty or invalid url"
		s.logger.Info("empty or invalid url")
		return &resp, nil
	}

	// add to db after validation
	const stmt = `insert into links (short, url, createdAt) values ($1, $2, $3)`
	_, err := s.db.Exec(ctx, stmt, shortCode, in.Input, time.Now())
	if err != nil {
		// TODO: handle postgres error with error types
		resp.Code = 500
		resp.Error = "internal server error"
		s.logger.Errorf("unable to insert url: %v and shortcode: %v -> error: %v", in.Input, shortCode, err)
		return &resp, nil
	}

	//build response
	message.Url = in.Input
	message.Id = shortCode
	resp.Code = 200
	resp.Message = &message
	s.logger.Infof("success: %v", &resp)
	return &resp, nil
}

// FetchUrl controls our grpc proto.FetchUrl service which accepts an id and returns the long url
func (s *urlShortenerServer) FetchUrl(ctx context.Context, in *proto.Request) (*proto.Response, error) {
	var resp proto.Response
	var message proto.Message

	const stmt = `select url from links where short=$1 limit 1`
	err := s.db.QueryRow(ctx, stmt, in.Input).Scan(&message.Url)
	if err != nil {
		// TODO: handle postgres error with error types
		resp.Code = 404
		resp.Error = "not found"
		s.logger.Errorf("unable to fetch shortcode: %v -> error: %v", in.Input, err)
		return &resp, nil
	}

	message.Id = in.Input
	resp.Code = 200
	resp.Message = &message
	s.logger.Infof("success: %v", &resp)
	return &resp, nil
}

// newServer is our server object constructor
//func newServer(db *pgxpool.Pool) *urlShortenerServer {
//	return &urlShortenerServer{db: db}
//}

// main function
func main() {
	flag.Parse()
	// start zap logger
	logger, _ := zap.NewProduction()
	sugarLog := logger.Sugar() // sugaredLogger from zap.SugaredLogger
	defer logger.Sync()        // flushes buffer, if any
	// pgx dbpool
	dbURL := "postgres://postgres:postgres@localhost:5432/postgres"
	dbpool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		sugarLog.Fatalf("Unable to create postgres connection pool: %v\n", err)
	}
	defer dbpool.Close()
	// shortId manager
	shortCode, err := shortid.New(1, shortid.DefaultABC, 7665)
	if err != nil {
		sugarLog.Fatalf("failed to start shortid generator: %v", err)
	}
	// start net listener
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		sugarLog.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			*certFile = data.Path("x509/server_cert.pem")
		}
		if *keyFile == "" {
			*keyFile = data.Path("x509/server_key.pem")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			sugarLog.Fatalf("Failed to generate credentials: %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterUrlShortenerServer(grpcServer, &urlShortenerServer{logger: sugarLog, db: dbpool, shortCode: shortCode})
	sugarLog.Info("Starting server on port: ", *port)
	grpcServer.Serve(lis)
}
