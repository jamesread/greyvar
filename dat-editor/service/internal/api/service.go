package api

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/jamesread/greyvar/dat-editor/service/buildinfo"
	"github.com/jamesread/greyvar/dat-editor/service/gen/deditor/v1/deditorv1connect"
	deditorv1 "github.com/jamesread/greyvar/dat-editor/service/gen/deditor/v1"
	"github.com/jamesread/greyvar/dat-editor/service/internal/config"
	"github.com/jamesread/greyvar/dat-editor/service/internal/paths"
)

type Service struct {
	cfg     *config.Config
	resRoot string
}

func New(cfg *config.Config, resRoot string) *Service {
	return &Service{cfg: cfg, resRoot: resRoot}
}

func (s *Service) Init(_ context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[deditorv1.InitResponse], error) {
	return connect.NewResponse(&deditorv1.InitResponse{
		Version:   buildinfo.Version,
		SiteTitle: s.cfg.SiteTitle,
		DatDir:    paths.DatDirLabel(),
		ResDir:    s.resRoot,
	}), nil
}

func BuildMux(svc *Service) *http.ServeMux {
	mux := http.NewServeMux()
	path, handler := deditorv1connect.NewDatEditorServiceHandler(svc)
	mux.Handle(path, handler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir(svc.resRoot))))

	if staticDir := findStaticDir(); staticDir != "" {
		mux.Handle("/", spaFileServer(staticDir))
	}

	return mux
}

func findStaticDir() string {
	for _, candidate := range []string{"../frontend/dist", "./frontend/dist", "/app/frontend/dist"} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			abs, err := filepath.Abs(candidate)
			if err == nil {
				return abs
			}
		}
	}
	return ""
}

func spaFileServer(root string) http.Handler {
	fs := http.FileServer(http.Dir(root))
	indexPath := filepath.Join(root, "index.html")
	rootAbs, _ := filepath.Abs(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		target := spaTarget(root, rootAbs, r.URL.Path)
		if target == indexPath {
			http.ServeFile(w, r, indexPath)
			return
		}
		fs.ServeHTTP(w, r)
	})
}

func spaTarget(root, rootAbs, urlPath string) string {
	indexPath := filepath.Join(root, "index.html")
	path := filepath.Clean(urlPath)
	if path == "/" || path == "." {
		return indexPath
	}
	fpath := filepath.Join(root, path)
	abs, err := filepath.Abs(fpath)
	if err != nil || !strings.HasPrefix(abs, rootAbs) {
		return indexPath
	}
	if info, err := os.Stat(fpath); err == nil && info.Mode().IsRegular() {
		return fpath
	}
	return indexPath
}
