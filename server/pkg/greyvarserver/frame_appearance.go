package greyvarserver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jamesread/greyvar/datlib/rigs"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

func loadRigs() *rigs.Set {
	set, errs := rigs.LoadSet()
	for _, err := range errs {
		log.Warnf("Cannot load a character rig: %v", err)
	}
	if len(set.BySize) == 0 {
		log.Warn("No character rigs loaded; players use the default look (set GREYVAR_RES_DIR)")
		return nil
	}
	for size, r := range set.BySize {
		log.Infof("Loaded %s rig %s (%d layers)", size, r.Name, len(r.Layers))
	}
	// Every size collides with the same footprint at its feet (rig.yml hitbox,
	// in cell coordinates), so one hitbox serves them all.
	if hb := set.Size("").Hitbox; hb[2] > 0 && hb[3] > 0 {
		playerHitbox = hitbox{hb[0], hb[1], hb[2], hb[3]}
		log.Infof("Player hitbox %v (default size %s)", hb, set.Default)
	}
	return set
}

// rigFor is the rig a look is drawn with (nil when no rigs are loaded).
func (s *serverInterface) rigFor(a *rigs.Appearance) *rigs.Rig {
	if s.rigs == nil {
		return nil
	}
	return s.rigs.For(a)
}

// defaultLook is a stable default look (default size) for a name.
func (s *serverInterface) defaultLook(seed string) rigs.Appearance {
	if s.rigs == nil {
		return rigs.Appearance{Parts: map[string]string{}, Palettes: map[string]string{}}
	}
	return s.rigs.DefaultLook(seed)
}

// knownOnly splits requested parts/palettes into those the rig has and the
// rest, so a look saved under another size or an older rig still applies.
func knownOnly(r *rigs.Rig, parts, palettes map[string]string) (map[string]string, map[string]string, []string) {
	n := r.Normalize(rigs.Appearance{Parts: parts, Palettes: palettes})
	okParts, okPalettes := map[string]string{}, map[string]string{}
	var dropped []string
	for k, v := range parts {
		if n.Parts[k] == v {
			okParts[k] = v
		} else {
			dropped = append(dropped, k+"="+v)
		}
	}
	for k, v := range palettes {
		if n.Palettes[k] == v {
			okPalettes[k] = v
		} else {
			dropped = append(dropped, k+"="+v)
		}
	}
	sort.Strings(dropped)
	return okParts, okPalettes, dropped
}

func appearanceToPB(a *rigs.Appearance) *pb.Appearance {
	if a == nil {
		return nil
	}
	out := &pb.Appearance{Rig: a.Rig, Parts: map[string]string{}, Palettes: map[string]string{}, HeldItem: a.HeldItem}
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
	if s.rigs == nil || rp.Entity == nil {
		reply("appearance: character rigs not loaded on server")
		return
	}

	base := s.defaultLook(rp.Username)
	if rp.Entity.Appearance != nil {
		base = *rp.Entity.Appearance
	}
	if size := req.GetRig(); size != "" {
		target, ok := s.rigs.BySize[size]
		if !ok {
			reply(fmt.Sprintf("appearance: unknown size %q (have %v)", size, rigs.Sizes))
			return
		}
		base = target.Normalize(base)
	}
	r := s.rigFor(&base)
	parts, palettes, dropped := knownOnly(r, req.GetParts(), req.GetPalettes())
	if len(dropped) > 0 && len(parts)+len(palettes) == 0 && req.GetRig() == "" && req.GetHeldItem() == "" {
		// Nothing valid to change (e.g. a typo): say why, change nothing.
		if _, err := r.Apply(base, req.GetParts(), req.GetPalettes()); err != nil {
			reply("appearance: " + err.Error())
		}
		return
	}
	if len(dropped) > 0 {
		reply(fmt.Sprintf("appearance: the %s rig has no %s; ignored", r.Size, strings.Join(dropped, ", ")))
	}
	next, err := r.Apply(r.Normalize(base), parts, palettes)
	if err != nil {
		reply("appearance: " + err.Error())
		return
	}
	// The held item must be in the player's inventory. A bad held item only
	// drops that part, so a restored look still applies after reconnecting
	// with an empty inventory.
	switch held := req.GetHeldItem(); held {
	case "":
	case "-":
		next.HeldItem = ""
	default:
		if _, ok := s.itemDefinitions[held]; !ok {
			reply(fmt.Sprintf("appearance: unknown item %q", held))
		} else if rp.inventoryCount(held) <= 0 {
			reply(fmt.Sprintf("appearance: you have no %s to hold", held))
		} else {
			next.HeldItem = held
		}
	}

	setAppearance(s, rp, next)
	reply("appearance: " + describeAppearance(next))
}

// setAppearance stores a look and tells every client that knows the player.
func setAppearance(s *serverInterface, rp *RemotePlayer, next rigs.Appearance) {
	rp.Entity.Appearance = &next
	change := &pb.EntityAppearanceChange{EntityId: rp.Entity.ServerId, Appearance: appearanceToPB(&next)}
	for _, p := range s.remotePlayers {
		if _, known := p.KnownEntities[rp.Entity.ServerId]; known {
			p.PendingAppearanceChanges = append(p.PendingAppearanceChanges, change)
		}
	}
}

// putAwayHeldItemIfGone clears the held item once it is no longer in the
// player's inventory (crafted away, used...). Called on every inventory change.
func putAwayHeldItemIfGone(s *serverInterface, rp *RemotePlayer) {
	if rp.Entity == nil || rp.Entity.Appearance == nil {
		return
	}
	held := rp.Entity.Appearance.HeldItem
	if held == "" || rp.inventoryCount(held) > 0 {
		return
	}
	next := *rp.Entity.Appearance
	next.HeldItem = ""
	setAppearance(s, rp, next)
}

func describeAppearance(a rigs.Appearance) string {
	var parts []string
	if a.Rig != "" {
		parts = append(parts, "size="+a.Rig)
	}
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
