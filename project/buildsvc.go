package project

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/chinglinwen/log"

	pb "wen/self-release/pkg/proto/build"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func Build(project, branch, env, commitid string) (b *buildsvc) {
	b = newBuildsvc(project, branch, env, commitid)
	if defaultBuildClient == nil {
		b.err = fmt.Errorf("buildsvc not inited")
		return
	}
	b.Build()
	return
}

type buildClient struct {
	pb.BuildsvcClient
}

type buildsvc struct {
	project, branch, env, commitid string

	client *buildClient
	out    chan string
	err    error
}

var defaultBuildClient *buildClient

func newBuildClient(addr string) *buildClient {
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

	return &buildClient{BuildsvcClient: client}
}

func newBuildsvc(project, branch, env, commitid string) *buildsvc {
	return &buildsvc{
		project:  project,
		branch:   branch,
		env:      env,
		commitid: commitid,
		client:   defaultBuildClient,
	}
}

func (b *buildsvc) GetOutput() (chan string, error) {
	return b.out, b.err
}

func (b *buildsvc) GetError() error {
	return b.err
}

func (b *buildsvc) Build() {
	log.Debug.Printf("start build for %v, branch: %v, env: %v, commitid: %v\n",
		b.project, b.branch, b.env, b.commitid)

	r := &pb.Request{
		Project:  b.project,
		Branch:   b.branch,
		Env:      b.env,
		Commitid: b.commitid,
	}

	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	stream, err := b.client.Build(context.TODO(), r)
	if err != nil {
		b.err = fmt.Errorf("rpc call failed: %v", err)
		return
	}

	b.out = make(chan string, 100) // increase to 500, will cause later deepcopy panic
	go func() {
		// defer cancel()
		defer close(b.out)
		log.Debug.Printf("start receive output...")
		output := &pb.Response{}
		var e error
		for {
			e = stream.RecvMsg(output)
			if e == io.EOF {
				log.Debug.Printf("output reached eof")
				break
			}
			if e != nil {
				err = fmt.Errorf("stream receive output err, %v", e)
				log.Debug.Println(err)
				b.err = e
				return
			}
			// log.Printf("%v", output.GetOutput())
			b.out <- output.GetOutput()
		}
		log.Debug.Printf("done of receive output for %v\n", b.project)
	}()
	log.Debug.Printf("made rpc call for %v, receiving output now...\n", b.project)
	return
}
