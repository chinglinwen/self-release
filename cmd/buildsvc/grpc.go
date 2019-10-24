package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	buildpkg "wen/self-release/cmd/buildsvc/build"
	pb "wen/self-release/pkg/proto/build"
	projectpkg "wen/self-release/project"

	"github.com/chinglinwen/log"
	prettyjson "github.com/hokaccha/go-prettyjson"
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

func pretty(prefix string, a interface{}) {
	out, _ := prettyjson.Marshal(a)
	fmt.Printf("%v: %s\n", prefix, out)
}

func validateRequest(r *pb.Request) (err error) {
	if r.Project == "" {
		return fmt.Errorf("project is empty")
	}
	if r.Branch == "" {
		return fmt.Errorf("branch is empty")
	}

	if r.Env == "" {
		return fmt.Errorf("env is empty")
	}

	if r.Commitid == "" {
		return fmt.Errorf("commitid is empty")
	}
	return
}
func (s *buildServer) Build(r *pb.Request, stream pb.Buildsvc_BuildServer) (err error) {

	pretty("start build:", r)

	if err = validateRequest(r); err != nil {
		log.Println(err)
		return
	}

	if err := stream.Send(&pb.Response{Output: "buildsvc clone or open project..."}); err != nil {
		err = fmt.Errorf("send stream err: %v", err)
		log.Println(err)
		return err
	}

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

	project, branch, env, commitid := r.Project, r.Branch, r.Env, r.Commitid

	p, err := projectpkg.NewProject(project, projectpkg.SetBranch(branch))
	if err != nil {
		log.Printf("NewProject err: %v\n", err)
		return
	}
	workdir, err := p.GetWorkDir()
	if err != nil {
		return
	}
	log.Printf("start building image for project: %v, branch: %v, env: %v, commitid: %v\n", project, branch, env, commitid)

	if err := stream.Send(&pb.Response{Output: "buildsvc start building..."}); err != nil {
		err = fmt.Errorf("send stream err: %v", err)
		log.Println(err)
		return err
	}

	b := buildpkg.NewBuilder(workdir, project, branch, env, commitid)
	out, err := b.Output()
	if err != nil {
		err = fmt.Errorf("build err: %v", err)
		log.Println(err)
		return
	}

	log.Debug.Printf("build is started, checking symbol exist\n")
	detector := "digest: sha256"
	var success bool

	if err := stream.Send(&pb.Response{Output: "output start ==="}); err != nil {
		err = fmt.Errorf("send stream err: %v", err)
		log.Println(err)
		return err
	}

	for out.Scan() {
		if err := b.GetError(); err != nil {
			err = fmt.Errorf("build err: %v", err)
			log.Println(err)
			return err
		}
		text := out.Text()
		if strings.Contains(text, detector) {
			success = true
		}
		if err := stream.Send(&pb.Response{Output: text}); err != nil {
			err = fmt.Errorf("send stream err: %v", err)
			log.Println(err)
			return err
		}
	}
	if err := stream.Send(&pb.Response{Output: "output end ==="}); err != nil {
		err = fmt.Errorf("send stream err: %v", err)
		log.Println(err)
		return err
	}
	if !success {
		err = fmt.Errorf("build image failed, checkout logs")
		log.Println(err)
		return
	}
	log.Println("build ok")

	return
}

// https://stackoverflow.com/questions/32840687/timeout-for-waitgroup-wait
// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
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
