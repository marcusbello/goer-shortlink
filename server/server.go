package main

import (
	"context"
	"flag"
	"fmt"
	"goer-shortlink/data"
	"goer-shortlink/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"net"
	"strings"
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
	//TODO: implement logger and database
}

// ShortLink controls our proto.ShortLink service it accepts input url and returns the short url
func (s *urlShortenerServer) ShortLink(ctx context.Context, in *proto.Request) (*proto.Response, error) {
	var resp proto.Response
	var message proto.Message
	if in.Input == "" {
		resp.Error = "empty url"
		return &resp, nil
	}
	if strings.HasPrefix(in.Input, "https://") {
		message.Url = in.Input
	} else {
		message.Url = "http://nohttpsdefault.url"
	}
	resp.Message = &message
	return &resp, nil
}

// FetchUrl controls our grpc proto.FetchUrl service which accepts an id and returns the long url
func (s *urlShortenerServer) FetchUrl(ctx context.Context, in *proto.Request) (*proto.Response, error) {
	var resp proto.Response
	var message proto.Message

	if in.Input == "abcde" {
		message.Url = "https://testing.url"
	}
	resp.Code = 200
	resp.Message = &message

	return &resp, nil
}

/*
//func newServer() *urlShortenerServer {
//	return &urlShortenerServer{}
//}
*/

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
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
			log.Fatalf("Failed to generate credentials: %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterUrlShortenerServer(grpcServer, &urlShortenerServer{})
	grpcServer.Serve(lis)
}
