/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"net"
	"os"
	"context"
	"os/signal"
	"syscall"
	"crypto/tls"
	"strings"
	"encoding/pem"


	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	pb "sigs.k8s.io/gateway-api/conformance/grpc-echo/grpcechoserver"
)

type echoServer struct {
	pb.UnimplementedGrpcEchoServer
	fullService string
	tls bool
}

func fullMethod(svc, method string) string {
	return fmt.Sprintf("/%s/%s", svc, method)
}

func (s *echoServer) fullMethod(method string) string {
	return fullMethod(s.fullService, method)
}

func (s *echoServer) doEcho(methodName string, ctx context.Context, in *pb.EchoRequest) (*pb.EchoResponse, error) {
	connectionType := "plaintext"
	if s.tls {
		connectionType = "TLS"
	}
	fmt.Printf("Received over %s: %v\n", connectionType, in)
	mdElems, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		msg := "Failed to retrieve metadata from incoming request.\n"
		fmt.Printf(msg)
		return nil, fmt.Errorf(msg)
	}
	authority := ""
	headers := []*pb.Header{}
	for k, vs := range mdElems {
		for _, v := range vs {
			if k == ":authority" {
				authority = v
			}
			headers = append(headers, &pb.Header{
				Key: k,
				Value: v,
			})
		}
	}
	resp := &pb.EchoResponse{
		Assertions: &pb.Assertions{
			FullyQualifiedMethod: s.fullMethod(methodName),
			Headers: headers,
			Authority: authority,
			Context: podContext,
		},
	}
	if s.tls {
		tlsAssertions := &pb.TLSAssertions{}
		p, ok := peer.FromContext(ctx)
		if !ok {
			msg := "Failed to retrieve auth info from request\n"
			fmt.Printf(msg)
			return nil, fmt.Errorf(msg)
		}
		tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
		if !ok {
			msg := "Failed to retrieve TLS info from request\n"
			fmt.Printf(msg)
			return nil, fmt.Errorf(msg)
		}
		switch tlsInfo.State.Version {
		case tls.VersionTLS13:
			tlsAssertions.Version = "TLSv1.3"
		case tls.VersionTLS12:
			tlsAssertions.Version = "TLSv1.2"
		case tls.VersionTLS11:
			tlsAssertions.Version = "TLSv1.1"
		case tls.VersionTLS10:
			tlsAssertions.Version = "TLSv1.0"
		}

		tlsAssertions.NegotiatedProtocol = tlsInfo.State.NegotiatedProtocol
		tlsAssertions.ServerName = tlsInfo.State.ServerName
		tlsAssertions.CipherSuite = tls.CipherSuiteName(tlsInfo.State.CipherSuite)

		// Convert peer certificates to PEM blocks.
		for _, c := range tlsInfo.State.PeerCertificates {
			var out strings.Builder
			pem.Encode(&out, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: c.Raw,
			})
			tlsAssertions.PeerCertificates = append(tlsAssertions.PeerCertificates, out.String())
		}
		resp.Assertions.TlsAssertions = tlsAssertions
	}
	return resp, nil
}


func (s *echoServer) Echo(ctx context.Context, in *pb.EchoRequest) (*pb.EchoResponse, error) {
	return s.doEcho("Echo", ctx, in)
}

func (s *echoServer) EchoTwo(ctx context.Context, in *pb.EchoRequest) (*pb.EchoResponse, error) {
	return s.doEcho("EchoTwo", ctx, in)
}

var podContext *pb.Context


func main() {
	podContext = &pb.Context{
		Namespace: 	os.Getenv("NAMESPACE"),
		Ingress:   	os.Getenv("INGRESS_NAME"),
		ServiceName:   	os.Getenv("SERVICE_NAME"),
		Pod:       	os.Getenv("POD_NAME"),
	}

	svcs := pb.File_grpcecho_proto.Services()
	svcd := svcs.ByName("GrpcEcho")
	if svcd == nil {
		fmt.Println("failed to look up service GrpcEcho.\n")
		os.Exit(1)
	}
	fullService := string(svcd.FullName())

	// Set up plaintext server.
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "3000"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", httpPort))
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		os.Exit(1)
	}
	s := grpc.NewServer()
	pb.RegisterGrpcEchoServer(s, &echoServer{fullService: fullService, tls: false})
	reflection.Register(s)

	fmt.Printf("plaintext server listening at %v\n", lis.Addr())

	go s.Serve(lis)

	if os.Getenv("TLS_SERVER_CERT") != "" && os.Getenv("TLS_SERVER_PRIVKEY") != "" {
		// Set up TLS server.
		httpsPort := os.Getenv("HTTPS_PORT")
		if httpsPort == "" {
			httpsPort = "8443"
		}
		creds, err := credentials.NewServerTLSFromFile(os.Getenv("TLS_SERVER_CERT"), os.Getenv("TLS_SERVER_PRIVKEY") )
		if err != nil {
			fmt.Printf("failed to create credentials: %v\n", err)
			os.Exit(1)
		}
		secureListener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", httpsPort))
		if err != nil {
			fmt.Printf("failed to listen: %v\n", err)
			os.Exit(1)
		}
		secureServer := grpc.NewServer(grpc.Creds(creds))
		pb.RegisterGrpcEchoServer(secureServer, &echoServer{fullService: fullService, tls: true})
		reflection.Register(secureServer)

		fmt.Printf("secure server listening at %v\n", secureListener.Addr())
		go secureServer.Serve(secureListener)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
