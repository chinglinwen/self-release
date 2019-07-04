package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	// "google.golang.org/grpc/credentials"
	// "google.golang.org/grpc/testdata"

	buildpkg "wen/self-release/cmd/buildsvc/build"
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
	r     *pb.Request
	cache map[string]bool
	// mu map[string]sync.Mutex // protects routeNotes
}

// // need to cache request to verify?
// func compareRequest(r1, r2 *pb.Request) bool {
// 	if r1.Project != r2.Project {
// 		return false
// 	}
// 	if r1.Branch != r2.Branch {
// 		return false
// 	}
// 	if r1.Env != r2.Env {
// 		return false
// 	}
// 	return true
// }

func (s *buildServer) Build(r *pb.Request, stream pb.Buildsvc_BuildServer) (err error) {
	log.Printf("start build %v:%v:%v", r.Project, r.Branch, r.Env)

	key := fmt.Sprintf("%v:%v:%v", r.Project, r.Branch, r.Env)
	if _, ok := s.cache[key]; ok {
		err = fmt.Errorf("request is already in build for %v, you may try later", key)
		log.Println(err)
		return
	}
	s.cache[key] = true
	defer func() {
		delete(s.cache, key)
	}()

	// log.Println("in build")
	// time.Sleep(10 * time.Second)
	// return

	project, branch, env := r.Project, r.Branch, r.Env

	p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	if err != nil {
		return
	}

	log.Printf("start building image for project: %v, branch: %v, env: %v\n", project, branch, env)

	var wg sync.WaitGroup
	wg.Add(1)

	out := make(chan string, 100)
	defer close(out)
	err = buildpkg.BuildStreamOutput(p.WorkDir, project, branch, env, out, wg)
	// e := p.Build(project, branch, env, out)
	if err != nil {
		err = fmt.Errorf("build err: %v", err)
		return
	}

	detector := "digest: sha256"
	var success bool

	// log.Printf("docker build outputs: %v", out)
	// scanner := bufio.NewScanner(strings.NewReader(out))
	// scanner.Split(bufio.ScanLines)
	for text := range out {
		if strings.Contains(text, detector) {
			success = true
		}
		if err := stream.Send(&pb.Response{Output: text}); err != nil {
			return err
		}
	}
	if !success {
		err = fmt.Errorf("build image failed, checkout logs")
		log.Println(err)
		return
	}
	log.Println("build ok")
	wg.Wait()
	log.Println("end of handle build output")
	return nil
}

func newServer() *buildServer {
	s := &buildServer{cache: make(map[string]bool)}
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
