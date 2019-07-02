package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"

	// "google.golang.org/grpc/credentials"
	// "google.golang.org/grpc/testdata"

	pb "wen/self-release/pkg/proto/build"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	"google.golang.org/grpc"
)

var (
	// tls        = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	// certFile   = flag.String("cert_file", "", "The TLS cert file")
	// keyFile    = flag.String("key_file", "", "The TLS key file")
	// jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	grpcport = flag.Int("grpcport", 10000, "The server port")
)

type buildServer struct {
	r  *pb.Request
	mu sync.Mutex // protects routeNotes
}

func (s *buildServer) Build(r *pb.Request, stream pb.Buildsvc_BuildServer) (err error) {
	if s.r != nil && reflect.DeepEqual(s.r, r) {
		err = fmt.Errorf("already in build")
		return
	}
	project, branch, env := r.Project, r.Branch, r.Env

	p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	if err != nil {
		return
	}

	log.Printf("start building image for project: %v, branch: %v, env: %v\n", project, branch, env)
	// out := make(chan string)
	out, err := p.BuildStreamOutput(project, branch, env)
	// e := p.Build(project, branch, env, out)
	if err != nil {
		err = fmt.Errorf("build err: %v", err)
		return
	}

	// log.Printf("docker build outputs: %v", out)
	// scanner := bufio.NewScanner(strings.NewReader(out))
	// scanner.Split(bufio.ScanLines)
	for text := range out {
		if err := stream.Send(&pb.Response{Output: text}); err != nil {
			return err
		}
	}
	log.Println("build done")
	return nil
}

func newServer() *buildServer {
	s := &buildServer{}
	return s
}

func runGRPC() {
	// flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcport))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	// if *tls {
	// 	if *certFile == "" {
	// 		*certFile = testdata.Path("server1.pem")
	// 	}
	// 	if *keyFile == "" {
	// 		*keyFile = testdata.Path("server1.key")
	// 	}
	// 	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	// 	if err != nil {
	// 		log.Fatalf("Failed to generate credentials %v", err)
	// 	}
	// 	opts = []grpc.ServerOption{grpc.Creds(creds)}
	// }
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterBuildsvcServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(tamird): point to merged gRPC code rather than a PR.
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

/*
	srv := &http.Server{
		Addr:    demoAddr,
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*demoKeyPair},
			NextProtos:   []string{"h2"},
		},
	}
	fmt.Printf("grpc on port: %d\n", port)
	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
*/