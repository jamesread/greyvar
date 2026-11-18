package greyvarserver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jamesread/greyvar/datlib/rigs"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

func loadPlayerRig() *rigs.Rig {
	path := rigs.PlayerRigPath()
	if path == "" {
		log.Warn("Cannot find res dir; players use the default look (set GREYVAR_RES_DIR)")
		return nil
	}
	r, err := rigs.Load(path)
	if err != nil {
		log.Warnf("Cannot load player rig %s: %v; players use the default look", path, err)
		return nil
	}
	log.Infof("Loaded player rig from %s (%d layers)", path, len(r.Layers))
	return r
}

func appearanceToPB(a *rigs.Appearance) *pb.Appearance {
	if a == nil {
		return nil
	}
	out := &pb.Appearance{Parts: map[string]string{}, Palettes: map[string]string{}, HeldItem: a.HeldItem}
	for k, v := range a.Parts {
		out.Parts[k] = v
	}
	for k, v := range a.Palettes {
		out.Palettes[k] = v
	}
	return out
}

// processAppearanceRequest changes the requesting player's look and tells
// every client that already knows the player entity.
func processAppearanceRequest(s *serverInterface, rp *RemotePlayer, req *pb.AppearanceRequest) {
	reply := func(msg string) { rp.PendingConsoleMessages = append(rp.PendingConsoleMessages, msg) }
	if s.playerRig == nil || rp.Entity == nil {
		reply("appearance: player rig not loaded on server")
		return
	}

	base := s.playerRig.Default(rp.Username)
	if rp.Entity.Appearance != nil {
		base = *rp.Entity.Appearance
	}
	next, err := s.playerRig.Apply(base, req.GetParts(), req.GetPalettes())
	if err != nil {
		reply("appearance: " + err.Error())
		return
	}
	switch held := req.GetHeldItem(); held {
	case "":
	case "-":
		next.HeldItem = ""
	default:
		if _, ok := s.itemDefinitions[held]; !ok {
			reply(fmt.Sprintf("appearance: unknown item %q", held))
			return
		}
		next.HeldItem = held
	}

	rp.Entity.Appearance = &next
	change := &pb.EntityAppearanceChange{EntityId: rp.Entity.ServerId, Appearance: appearanceToPB(&next)}
	for _, p := range s.remotePlayers {
		if _, known := p.KnownEntities[rp.Entity.ServerId]; known {
			p.PendingAppearanceChanges = append(p.PendingAppearanceChanges, change)
		}
	}
	reply("appearance: " + describeAppearance(next))
}

func describeAppearance(a rigs.Appearance) string {
	var parts []string
	for k, v := range a.Parts {
		parts = append(parts, k+"="+v)
	}
	for k, v := range a.Palettes {
		parts = append(parts, k+"="+v)
	}
	sort.Strings(parts)
	if a.HeldItem != "" {
		parts = append(parts, "held="+a.HeldItem)
	}
	return strings.Join(parts, " ")
}

func frameAppearanceChanges(s *serverInterface, p *RemotePlayer) {
	if len(p.PendingAppearanceChanges) == 0 {
		return
	}
	p.currentFrame.EntityAppearanceChanges = append(p.currentFrame.EntityAppearanceChanges, p.PendingAppearanceChanges...)
	p.PendingAppearanceChanges = nil
}
