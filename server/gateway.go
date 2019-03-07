package server

import (
	pb "fabric-sdk-go/protos"
	"fabric-sdk-go/server/helpers"
	"fabric-sdk-go/server/services"
	"fabric-sdk-go/third_party/swagger-ui"
	"net"
	"net/http"
	"path"
	"strings"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	ServerPort string
	SwaggerDir string
	EndPoint   string
)

var logger = helpers.GetLogger()

func Run() (err error) {
	EndPoint = ":" + ServerPort
	conn, err := net.Listen("tcp", EndPoint)
	if err != nil {
		logger.Error("TCP Listen err:%s", err)
	}

	services.Init()

	//srv := newServer()
	srv := newGrpc()
	logger.Info("gRPC and https listen on: %s", ServerPort)

	if err = srv.Serve(conn); err != nil {
		logger.Error("ListenAndServe: %s", err)
	}

	return err
}

func newServer() *http.Server {
	grpcServer := newGrpc()
	gwmux, err := newGateway()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)
	mux.HandleFunc("/swagger/", serveSwaggerFile)
	serveSwaggerUI(mux)

	return &http.Server{
		Addr:    EndPoint,
		Handler: grpcHandlerFunc(grpcServer, mux),
	}
}

func newGrpc() *grpc.Server {
	server := grpc.NewServer()
	// TODO
	pb.RegisterChannelServer(server, services.NewChannelService())

	return server
}

func newGateway() (http.Handler, error) {
	ctx := context.Background()
	dopts := []grpc.DialOption{grpc.WithInsecure()}
	gwmux := runtime.NewServeMux()
	// TODO
	if err := pb.RegisterChannelHandlerFromEndpoint(ctx, gwmux, EndPoint, dopts); err != nil {
		return nil, err
	}

	return gwmux, nil
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	if otherHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			grpcServer.ServeHTTP(w, r)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func serveSwaggerFile(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "swagger.json") {
		logger.Error("Not Found: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join(SwaggerDir, p)

	logger.Info("Serving swagger-file: %s", p)

	http.ServeFile(w, r, p)
}

func serveSwaggerUI(mux *http.ServeMux) {
	fileServer := http.FileServer(&assetfs.AssetFS{
		Asset:    swagger.Asset,
		AssetDir: swagger.AssetDir,
		Prefix:   "third_party/swagger-ui",
	})
	prefix := "/swagger-ui/"
	mux.Handle(prefix, http.StripPrefix(prefix, fileServer))
}