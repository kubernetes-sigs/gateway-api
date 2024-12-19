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

package grpc

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"

	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
)

type serverConfig struct {
	// Controlled by HTTP_PORT env var
	HTTPPort int

	// Controlled by multiple env vars -- one for each field:
	//   - NAMESPACE
	//   - INGRESS_NAME
	//   - SERVICE_NAME
	//   - POD_NAME
	PodContext *pb.Context

	// Controlled by TLS_SERVER_CERT env var
	TLSServerCert string

	// Controlled by TLS_SERVER_PRIVKEY env var
	TLSServerPrivKey string

	// Controlled by HTTPS_PORT env var
	HTTPSPort int
}

type echoServer struct {
	pb.UnimplementedGrpcEchoServer
	fullService string
	tls         bool
	podContext  *pb.Context
}

func fullMethod(svc, method string) string {
	return fmt.Sprintf("/%s/%s", svc, method)
}

func (s *echoServer) fullMethod(method string) string {
	return fullMethod(s.fullService, method)
}

func (s *echoServer) doEcho(methodName string, ctx context.Context, in *pb.EchoRequest) (*pb.EchoResponse, error) { //nolint:revive // Method signature is determined by gRPC.
	connectionType := "plaintext"
	if s.tls {
		connectionType = "TLS"
	}
	fmt.Printf("Received over %s: %v\n", connectionType, in)
	mdElems, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		msg := "failed to retrieve metadata from incoming request"
		fmt.Println(msg)
		return nil, errors.New(msg)
	}
	authority := ""
	headers := []*pb.Header{}
	for k, vs := range mdElems {
		for _, v := range vs {
			if k == ":authority" {
				authority = v
			}
			headers = append(headers, &pb.Header{
				Key:   k,
				Value: v,
			})
		}
	}
	resp := &pb.EchoResponse{
		Assertions: &pb.Assertions{
			FullyQualifiedMethod: s.fullMethod(methodName),
			Headers:              headers,
			Authority:            authority,
			Context:              s.podContext,
		},
	}
	if s.tls {
		// TODO: Pull this out into a function so that we can unit test it.
		tlsAssertions := &pb.TLSAssertions{}
		p, ok := peer.FromContext(ctx)
		if !ok {
			msg := "failed to retrieve auth info from request"
			fmt.Println(msg)
			return nil, errors.New(msg)
		}
		tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
		if !ok {
			msg := "failed to retrieve TLS info from request"
			fmt.Println(msg)
			return nil, errors.New(msg)
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
			err := pem.Encode(&out, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: c.Raw,
			})
			if err != nil {
				fmt.Printf("failed to encode certificate: %v\n", err)
			} else {
				tlsAssertions.PeerCertificates = append(tlsAssertions.PeerCertificates, out.String())
			}
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

func runServer(config serverConfig) (int, int) { //nolint:unparam
	svcs := pb.File_grpcecho_proto.Services()
	svcd := svcs.ByName("GrpcEcho")
	if svcd == nil {
		fmt.Println("failed to look up service GrpcEcho.")
		os.Exit(1)
	}
	fullService := string(svcd.FullName())

	// Set up plaintext server.
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.HTTPPort))
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		os.Exit(1)
	}
	resolvedHTTPPort := lis.Addr().(*net.TCPAddr).Port
	s := grpc.NewServer()
	pb.RegisterGrpcEchoServer(s, &echoServer{fullService: fullService, tls: false, podContext: config.PodContext})
	reflection.Register(s)

	fmt.Printf("plaintext server listening at %v\n", lis.Addr())

	go func() {
		if err := s.Serve(lis); err != nil {
			fmt.Printf("failed to serve: %v\n", err)
			os.Exit(1)
		}
	}()

	resolvedHTTPSPort := -1
	if config.TLSServerCert != "" && config.TLSServerPrivKey != "" {
		// Set up TLS server.
		creds, err := credentials.NewServerTLSFromFile(config.TLSServerCert, config.TLSServerPrivKey)
		if err != nil {
			fmt.Printf("failed to create credentials: %v\n", err)
			os.Exit(1)
		}
		secureListener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.HTTPSPort))
		if err != nil {
			fmt.Printf("failed to listen: %v\n", err)
			os.Exit(1)
		}
		resolvedHTTPSPort = secureListener.Addr().(*net.TCPAddr).Port
		secureServer := grpc.NewServer(grpc.Creds(creds))
		pb.RegisterGrpcEchoServer(secureServer, &echoServer{fullService: fullService, tls: true, podContext: config.PodContext})
		reflection.Register(secureServer)

		fmt.Printf("secure server listening at %v\n", secureListener.Addr())
		go func() {
			err := secureServer.Serve(secureListener)
			if err != nil {
				fmt.Printf("failed to serve: %v\n", err)
				os.Exit(1)
			}
		}()
	}

	return resolvedHTTPPort, resolvedHTTPSPort
}

func Main() {
	podContext := &pb.Context{
		Namespace:   os.Getenv("NAMESPACE"),
		Ingress:     os.Getenv("INGRESS_NAME"),
		ServiceName: os.Getenv("SERVICE_NAME"),
		Pod:         os.Getenv("POD_NAME"),
	}
	var err error
	httpPortStr := os.Getenv("HTTP_PORT")
	var httpPort int
	if httpPortStr == "" {
		httpPort = 3000
	} else {
		httpPort, err = strconv.Atoi(httpPortStr)
		if err != nil {
			fmt.Printf("non-integer value in HTTP_PORT '%s': %v\n", httpPortStr, err)
			os.Exit(1)
		}
	}

	httpsPortStr := os.Getenv("HTTPS_PORT")
	var httpsPort int
	if httpsPortStr == "" {
		httpsPort = 8443
	} else {
		httpsPort, err = strconv.Atoi(httpsPortStr)
		if err != nil {
			fmt.Printf("non-integer value in HTTPS_PORT '%s': %v\n", httpsPortStr, err)
			os.Exit(1)
		}
	}

	config := serverConfig{
		HTTPPort:         httpPort,
		PodContext:       podContext,
		TLSServerCert:    os.Getenv("TLS_SERVER_CERT"),
		TLSServerPrivKey: os.Getenv("TLS_SERVER_PRIV_KEY"),
		HTTPSPort:        httpsPort,
	}
	runServer(config)
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
