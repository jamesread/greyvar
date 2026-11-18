package api

import (
	"context"
	"regexp"
	"sort"
	"strings"

	"connectrpc.com/connect"
	"github.com/jamesread/greyvar/datlib/entdefs"
	"google.golang.org/protobuf/types/known/emptypb"

	deditorv1 "github.com/jamesread/greyvar/dat-editor/service/gen/deditor/v1"
)

var entdefNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

func (s *Service) ListEntdefs(_ context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[deditorv1.ListEntdefsResponse], error) {
	names, err := entdefs.ListEntdefs()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	sort.Strings(names)
	items := make([]*deditorv1.EntdefSummary, 0, len(names))
	for _, name := range names {
		def, err := entdefs.ReadEntdef(name)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		items = append(items, &deditorv1.EntdefSummary{
			Name:         name,
			Title:        def.Title,
			InitialState: def.InitialState,
			Solid:        def.Solid,
			StateCount:   int32(len(def.States)),
		})
	}

	return connect.NewResponse(&deditorv1.ListEntdefsResponse{Items: items}), nil
}

func (s *Service) GetEntdef(_ context.Context, req *connect.Request[deditorv1.GetEntdefRequest]) (*connect.Response[deditorv1.Entdef], error) {
	def, err := entdefs.ReadEntdef(req.Msg.Name)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(entdefToProto(req.Msg.Name, def)), nil
}

func (s *Service) CreateEntdef(_ context.Context, req *connect.Request[deditorv1.CreateEntdefRequest]) (*connect.Response[deditorv1.NameResponse], error) {
	name := strings.TrimSpace(req.Msg.Name)
	if !entdefNamePattern.MatchString(name) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errInvalidEntdefName)
	}
	if _, err := entdefs.ReadEntdef(name); err == nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, errEntdefExists)
	}

	def, err := entdefFromCreate(req.Msg, name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := entdefs.WriteEntdef(name, def); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&deditorv1.NameResponse{Name: name}), nil
}

func (s *Service) UpdateEntdef(_ context.Context, req *connect.Request[deditorv1.UpdateEntdefRequest]) (*connect.Response[deditorv1.NameResponse], error) {
	def, err := entdefFromUpdate(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := entdefs.WriteEntdef(req.Msg.Name, def); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&deditorv1.NameResponse{Name: req.Msg.Name}), nil
}

func (s *Service) DeleteEntdef(_ context.Context, req *connect.Request[deditorv1.DeleteEntdefRequest]) (*connect.Response[deditorv1.DeleteResponse], error) {
	if err := entdefs.DeleteEntdef(req.Msg.Name); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&deditorv1.DeleteResponse{Deleted: req.Msg.Name}), nil
}

func entdefToProto(name string, def *entdefs.EntityDefinition) *deditorv1.Entdef {
	states := map[string]*deditorv1.EntityState{}
	for key, state := range def.States {
		states[key] = &deditorv1.EntityState{Name: state.Name}
	}
	return &deditorv1.Entdef{
		Name:         name,
		Title:        def.Title,
		InitialState: def.InitialState,
		Solid:        def.Solid,
		States:       states,
	}
}

func entdefFromCreate(msg *deditorv1.CreateEntdefRequest, name string) (*entdefs.EntityDefinition, error) {
	return buildEntdef(name, msg.Title, msg.InitialState, msg.Solid, msg.States)
}

func entdefFromUpdate(msg *deditorv1.UpdateEntdefRequest) (*entdefs.EntityDefinition, error) {
	return buildEntdef(msg.Name, msg.Title, msg.InitialState, msg.Solid, msg.States)
}

func buildEntdef(name, title, initialState string, solid bool, states map[string]*deditorv1.EntityState) (*entdefs.EntityDefinition, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		title = name
	}

	outStates := map[string]entdefs.EntityState{}
	for key, state := range states {
		display := state.Name
		if display == "" {
			display = key
		}
		outStates[key] = entdefs.EntityState{Name: display}
	}
	if outStates == nil {
		outStates = map[string]entdefs.EntityState{}
	}

	return &entdefs.EntityDefinition{
		Title:        title,
		InitialState: strings.TrimSpace(initialState),
		Solid:        solid,
		States:       outStates,
	}, nil
}
