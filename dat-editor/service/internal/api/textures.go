package api

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"connectrpc.com/connect"

	deditorv1 "github.com/jamesread/greyvar/dat-editor/service/gen/deditor/v1"
)

func (s *Service) ListTextures(_ context.Context, req *connect.Request[deditorv1.ListTexturesRequest]) (*connect.Response[deditorv1.ListTexturesResponse], error) {
	subdir, err := textureSubdir(req.Msg.Kind)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if s.resRoot == "" {
		return connect.NewResponse(&deditorv1.ListTexturesResponse{}), nil
	}

	dir := filepath.Join(s.resRoot, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return connect.NewResponse(&deditorv1.ListTexturesResponse{}), nil
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".png") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	return connect.NewResponse(&deditorv1.ListTexturesResponse{Filenames: files}), nil
}

func textureSubdir(kind string) (string, error) {
	switch kind {
	case "tiles":
		return filepath.Join("img", "textures", "tiles"), nil
	case "entities":
		return filepath.Join("img", "textures", "entities"), nil
	default:
		return "", errUnknownTextureKind
	}
}
