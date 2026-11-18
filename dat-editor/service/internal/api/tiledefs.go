package api

import (
	"context"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/jamesread/greyvar/datlib/texdefs"
	"google.golang.org/protobuf/types/known/emptypb"

	deditorv1 "github.com/jamesread/greyvar/dat-editor/service/gen/deditor/v1"
)

var defaultTiledefs = []*deditorv1.TiledefSummary{
	{Name: "water", Texture: "water.png", Traversable: false, HasFile: false},
	{Name: "barrier", Texture: "barrier.png", Traversable: false, HasFile: false},
}

func (s *Service) ListTiledefs(_ context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[deditorv1.ListTiledefsResponse], error) {
	byName := map[string]*deditorv1.TiledefSummary{}
	for _, def := range defaultTiledefs {
		byName[def.Name] = def
	}

	loaded, err := texdefs.ListTiledefs()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	for _, def := range loaded {
		name := strings.TrimSuffix(def.Filename, ".yaml")
		traversable := true
		if def.Traversable != nil {
			traversable = *def.Traversable
		}
		byName[name] = &deditorv1.TiledefSummary{
			Name:        name,
			Texture:     def.TextureName,
			Traversable: traversable,
			HasFile:     true,
		}
	}

	items := make([]*deditorv1.TiledefSummary, 0, len(byName))
	for _, item := range byName {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return connect.NewResponse(&deditorv1.ListTiledefsResponse{Items: items}), nil
}

func (s *Service) GetTiledef(_ context.Context, req *connect.Request[deditorv1.GetTiledefRequest]) (*connect.Response[deditorv1.Tiledef], error) {
	def, err := texdefs.ReadTiledef(req.Msg.Name)
	if err != nil {
		for _, fallback := range defaultTiledefs {
			if fallback.Name == req.Msg.Name {
				return connect.NewResponse(&deditorv1.Tiledef{
					Name:        fallback.Name,
					Texture:     fallback.Texture,
					Traversable: fallback.Traversable,
					HasFile:     false,
				}), nil
			}
		}
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	traversable := true
	if def.Traversable != nil {
		traversable = *def.Traversable
	}

	return connect.NewResponse(&deditorv1.Tiledef{
		Name:        strings.TrimSuffix(def.Filename, ".yaml"),
		Texture:     def.TextureName,
		Traversable: traversable,
		HasFile:     true,
	}), nil
}

func (s *Service) CreateTiledef(_ context.Context, req *connect.Request[deditorv1.CreateTiledefRequest]) (*connect.Response[deditorv1.NameResponse], error) {
	name := strings.TrimSpace(req.Msg.Name)
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errTiledefNameRequired)
	}
	def := tiledefFromRequest(req.Msg.Texture, req.Msg.Traversable)
	if err := texdefs.WriteTiledef(name, def); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&deditorv1.NameResponse{Name: name}), nil
}

func (s *Service) UpdateTiledef(_ context.Context, req *connect.Request[deditorv1.UpdateTiledefRequest]) (*connect.Response[deditorv1.NameResponse], error) {
	def := tiledefFromRequest(req.Msg.Texture, req.Msg.Traversable)
	if err := texdefs.WriteTiledef(req.Msg.Name, def); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&deditorv1.NameResponse{Name: req.Msg.Name}), nil
}

func (s *Service) DeleteTiledef(_ context.Context, req *connect.Request[deditorv1.DeleteTiledefRequest]) (*connect.Response[deditorv1.DeleteResponse], error) {
	if err := texdefs.DeleteTiledef(req.Msg.Name); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&deditorv1.DeleteResponse{Deleted: req.Msg.Name}), nil
}

func tiledefFromRequest(texture string, traversable bool) *texdefs.TileDefinition {
	return &texdefs.TileDefinition{
		Texture:     strings.TrimSpace(texture),
		Traversable: &traversable,
	}
}
