package main

import (
	"context"
	"flag"
	"io"
	"log"

	pb "wen/self-release/pkg/proto/build"

	"google.golang.org/grpc"
)

var (
	// tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	// caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

func build(client pb.BuildsvcClient, r *pb.Request) {
	log.Printf("start build %v", r)
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	// defer cancel()
	stream, err := client.Build(context.TODO(), r)
	if err != nil {
		log.Fatalf("%v.build err %v", client, err)
	}
	out := &pb.Response{}
	for {
		// out, err := stream.Recv()
		err := stream.RecvMsg(out)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.build output err, %v", client, err)
		}
		log.Printf("%v", out.GetOutput())
	}
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
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
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewBuildsvcClient(conn)

	build(client, &pb.Request{
		Project:  "robot/main",
		Branch:   "develop",
		Env:      "test",
		Commitid: "076f2793",
	})
}
