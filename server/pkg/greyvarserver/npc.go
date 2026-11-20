package greyvarserver

import (
	"math/rand"
	"time"

	"github.com/jamesread/greyvar/datlib/entdefs"
	"github.com/jamesread/greyvar/datlib/rigs"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

// NPCs are entities whose entdef has an appearance (drawn with the player
// rig), wander (stroll side to side near home) and/or dialog (see dialog.go).

const (
	defaultNPCSpeedPx = 3
	// talkPauseMs keeps an NPC still (facing the player) while you talk.
	talkPauseMs = 6000
)

var npcRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type npcState struct {
	homeX, homeY int32
	walking      bool
	targetX      int32
	waitUntil    int64 // ms; stand still until then
	nextStep     int64 // ms; wander steps are timed, not per frame
}

// npcStepMs: wander.speedPx is per this many ms.
const npcStepMs = 100

// setupNPC gives a freshly placed entity its fixed look and wander state.
func (s *serverInterface) setupNPC(ent *Entity, entdef *entdefs.EntityDefinition) {
	if entdef.Appearance != nil {
		ent.Appearance = s.npcAppearance(ent.Definition, entdef.Appearance)
		s.dressFigure(ent) // a mannequin's outfit (map property `wear`)
	}
	if entdef.Wander != nil && entdef.Wander.RangePx > 0 {
		ent.npc = &npcState{homeX: ent.X, homeY: ent.Y, waitUntil: nowMs() + s.npcPause(entdef.Wander)}
	}
}

func (s *serverInterface) npcAppearance(name string, look *entdefs.Look) *rigs.Appearance {
	if s.rigs == nil {
		return &rigs.Appearance{Rig: look.Rig, Parts: look.Parts, Palettes: look.Palettes, HeldItem: look.HeldItem}
	}
	// NPCs are adults unless their look says otherwise (rig: child).
	size := look.Rig
	if size == "" {
		size = "adult"
	}
	r := s.rigs.Size(size)
	base := r.Default(name)
	parts, palettes, dropped := knownOnly(r, look.Parts, look.Palettes)
	if len(dropped) > 0 {
		log.Warnf("entdef %q appearance: the %s rig has no %v; using its default for those", name, r.Size, dropped)
	}
	a, err := r.Apply(base, parts, palettes)
	if err != nil {
		log.Warnf("entdef %q appearance: %v; using the rig default", name, err)
		a = base
	}
	a.HeldItem = look.HeldItem
	return &a
}

func (s *serverInterface) npcPause(w *entdefs.Wander) int64 {
	if w == nil || len(w.PauseMs) < 2 || w.PauseMs[1] <= w.PauseMs[0] {
		if w != nil && len(w.PauseMs) > 0 {
			return w.PauseMs[0]
		}
		return 2000
	}
	return w.PauseMs[0] + npcRand.Int63n(w.PauseMs[1]-w.PauseMs[0]+1)
}

// tickNPCs moves wandering NPCs, but only on grids where a player is: nobody
// sees the others, so they cost nothing.
func (s *serverInterface) tickNPCs(now int64) {
	occupied := map[[2]string]bool{}
	for _, p := range s.remotePlayers {
		occupied[[2]string{p.CurrentWorldId, p.CurrentGridId}] = true
	}
	for _, ent := range s.entityInstances {
		if ent.npc == nil || !occupied[[2]string{ent.WorldId, ent.GridId}] || now < ent.npc.waitUntil || now < ent.npc.nextStep {
			continue
		}
		ent.npc.nextStep = now + npcStepMs
		entdef := s.entityDefinitions[ent.Definition]
		if entdef == nil || entdef.Wander == nil {
			continue
		}
		s.stepNPC(ent, entdef.Wander, now)
	}
}

func (s *serverInterface) stepNPC(ent *Entity, w *entdefs.Wander, now int64) {
	st := ent.npc
	speed := w.SpeedPx
	if speed <= 0 {
		speed = defaultNPCSpeedPx
	}
	if !st.walking {
		// Side to side only: a random spot within range of home, in steps.
		st.targetX = st.homeX + (npcRand.Int31n(2*w.RangePx/speed+1)-w.RangePx/speed)*speed
		if absInt32(st.targetX-ent.X) < speed {
			st.waitUntil = now + s.npcPause(w)
			return
		}
		st.walking = true
	}
	dx := speed
	if st.targetX < ent.X {
		dx = -speed
	}
	if absInt32(st.targetX-ent.X) < speed {
		dx = st.targetX - ent.X
	}
	nx := ent.X + dx
	if !s.npcCanStand(ent, nx, ent.Y) {
		st.walking = false
		st.waitUntil = now + s.npcPause(w)
		return
	}
	ent.X = nx
	if ent.X == st.targetX {
		st.walking = false
		st.waitUntil = now + s.npcPause(w)
	}
}

// npcCanStand: the NPC (player-sized) fits at (x, y): tiles, solid entities
// and players do not block it.
func (s *serverInterface) npcCanStand(ent *Entity, x, y int32) bool {
	if !s.isTraversableOnGrid(ent.WorldId, ent.GridId, x, y) {
		return false
	}
	if blocked, _ := s.solidEntityBlocks(ent.WorldId, ent.GridId, ent, x, y); blocked {
		return false
	}
	hb := playerHitbox
	for _, p := range s.remotePlayers {
		pe := p.Entity
		if pe == nil || p.CurrentWorldId != ent.WorldId || p.CurrentGridId != ent.GridId {
			continue
		}
		if absInt32(pe.X-x) < hb.W && absInt32(pe.Y-y) < hb.H {
			return false
		}
	}
	return true
}

// faceTowards stops a wandering NPC and turns it to face the player: a
// one-off "face<Dir>" event (clients turn it and keep it idle).
func (s *serverInterface) faceTowards(ent *Entity, rp *RemotePlayer) {
	if ent.npc != nil {
		ent.npc.walking = false
		ent.npc.waitUntil = nowMs() + talkPauseMs
	}
	if ent.Appearance == nil || rp.Entity == nil {
		return
	}
	dx, dy := rp.Entity.X-ent.X, rp.Entity.Y-ent.Y
	dir := "Down"
	switch {
	case absInt32(dx) >= absInt32(dy) && dx < 0:
		dir = "Left"
	case absInt32(dx) >= absInt32(dy):
		dir = "Right"
	case dy < 0:
		dir = "Up"
	}
	change := &pb.EntityStateChange{EntityId: ent.ServerId, NewState: "face" + dir}
	for _, p := range s.remotePlayers {
		if p.currentFrame == nil || p.CurrentWorldId != ent.WorldId || p.CurrentGridId != ent.GridId {
			continue
		}
		if _, known := p.KnownEntities[ent.ServerId]; known {
			p.currentFrame.EntityStateChanges = append(p.currentFrame.EntityStateChanges, change)
		}
	}
}
