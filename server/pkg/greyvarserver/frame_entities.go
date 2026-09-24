package greyvarserver

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"

	datlib "github.com/jamesread/greyvar/datlib/common"
	"github.com/jamesread/greyvar/datlib/tiled"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func frameNewEntdefs(s *serverInterface, p *RemotePlayer) {
	for name := range s.entityDefinitions {
		if _, ok := p.KnownEntdefs[name]; ok {
			continue
		}

		log.Infof("Need to tell %v about entdef %v", p.Username, name)

		serverEntdef := s.entityDefinitions[name]
		title := serverEntdef.Title
		if title == "" {
			title = name
		}

		netEntdef := &pb.EntityDefinition{
			Name: title,
		}

		for stateName := range serverEntdef.States {
			vis, ok := s.entityVisuals.ResolveState(title, stateName, serverEntdef.InitialState)
			if !ok {
				log.Warnf("entdef %q state %q has no tileset visual (including %s)", title, stateName, tiled.DefaultEntityVisualType)
				netEntdef.States = append(netEntdef.States, &pb.EntityState{Name: stateName})
				continue
			}
			if vis.Type == tiled.DefaultEntityVisualType && title != tiled.DefaultEntityVisualType {
				log.Infof("entdef %q state %q using default visual %s", title, stateName, tiled.DefaultEntityVisualType)
			}

			if netEntdef.Texture == "" {
				netEntdef.Texture = vis.Texture
			}

			netState := &pb.EntityState{
				Name:   stateName,
				Frames: append([]int32(nil), vis.Frames...),
				Holds:  append([]int32(nil), vis.Holds...),
			}
			for _, rect := range vis.Collision {
				netState.Collision = append(netState.Collision, &pb.CollisionRect{
					X: float32(rect.X),
					Y: float32(rect.Y),
					W: float32(rect.W),
					H: float32(rect.H),
				})
			}

			netEntdef.States = append(netEntdef.States, netState)
		}

		p.currentFrame.EntityDefinitions = append(p.currentFrame.EntityDefinitions, netEntdef)
		p.KnownEntdefs[name] = true
	}
}

func loadEntityVisuals() *tiled.EntityVisualCatalog {
	dir := filepath.Join(datlib.DatDir(), "entdefs")
	cat, err := tiled.LoadEntityVisualCatalog(dir)
	if err != nil {
		log.Warnf("Cannot load entity tileset visuals from %s: %v", dir, err)
		return &tiled.EntityVisualCatalog{}
	}
	log.Infof("Loaded entity tileset visuals from %s (%d types)", dir, cat.TypeCount())
	return cat
}
