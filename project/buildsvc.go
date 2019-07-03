package project

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	pb "wen/self-release/pkg/proto/build"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
// tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
// caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
// serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
// buildsvcAddr = flag.String("buildsvc", "buildsvc", "buildsvc address host:port ( or k8s service name )")
)

// func build(client pb.BuildsvcClient, r *pb.Request) (err error) {
// 	log.Printf("reqesting... %v", r)
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
// 	defer cancel()
// 	stream, err := client.Build(ctx, r)
// 	if err != nil {
// 		err = fmt.Errorf("build for %v err:", r.Project, err)
// 		return
// 	}
// 	for {
// 		out, err := stream.Recv()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			log.Fatalf("%v.build output err, %v", client, err)
// 		}
// 		log.Printf("%v", out.GetOutput())
// 	}
// 	return
// }

type Buildsvc struct {
	client pb.BuildsvcClient
}

var defaultBuildsvc *Buildsvc

func NewBuildSVC(addr string) *Buildsvc {
	log.Printf("connect buildsvc with: %v\n", addr)

	// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/retry/examples_test.go
	retryopts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted),
	}

	opts := []grpc.DialOption{
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryopts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryopts...)),
	}
	// if *tls {
	// 	if *caFile == "" {
	// 		*caFile = testdata.Path("ca.pem")
	// 	}
	// 	creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
	// 	if err != nil {
	// 		log.Fatalf("Failed to create TLS credentials %v", err)
	// 	}
	// 	opts = append(opts, grpc.WithTransportCredentials(creds))
	// } else {
	opts = append(opts, grpc.WithInsecure())
	// }
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Fatalf("fail to dial buildsvc, err: %v", err)
	}
	// defer conn.Close() // we don't close, unless program stopped
	client := pb.NewBuildsvcClient(conn)

	return &Buildsvc{client: client}
}

func (b *Buildsvc) Build(project, branch, env string) (out chan string, err error) {
	out = make(chan string)
	defer close(out)

	r := &pb.Request{
		Project: project,
		Branch:  branch,
		Env:     env,
	}

	log.Printf("reqesting... %v", r)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	stream, err := b.client.Build(ctx, r)
	if err != nil {
		err = fmt.Errorf("project: %v, build err: %v", project, err)
		return
	}
	for {
		output, e := stream.Recv()
		if e == io.EOF {
			break
		}
		if e != nil {
			err = fmt.Errorf("project: %v, receive output err, %v", project, e)
			return
		}
		// log.Printf("%v", out.GetOutput())
		out <- output.GetOutput()
	}
	return
}

func Build(project, branch, env string) (out chan string, err error) {
	if defaultBuildsvc == nil {
		err = fmt.Errorf("buildsvc not inited")
		return
	}
	return defaultBuildsvc.Build(project, branch, env)
}
