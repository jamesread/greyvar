package greyvarserver

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jamesread/greyvar/datlib/entdefs"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

// Puzzle wiring is data: Tiled object properties on map entities.
//
//	id          name other objects use to refer to this one (e.g. sw1)
//	loot        "item[:count], ..." granted when harvested/opened (overrides entdef loot)
//	requires    item consumed to open (interact.open entities, e.g. a locked chest)
//	msg         text shown when opened (and for signs, on touch)
//	locked_msg  text shown when `requires` is missing
//	open_when   "id=state, id=state": this entity is "open" while all hold,
//	            otherwise "closed" (e.g. a cell door watching levers)

const (
	propID        = "id"
	propLoot      = "loot"
	propRequires  = "requires"
	propMsg       = "msg"
	propLockedMsg = "locked_msg"
	propOpenWhen  = "open_when"
)

// entityID is the map object's puzzle id (its `id` property).
func entityID(ent *Entity) string {
	if ent == nil {
		return ""
	}
	return strings.TrimSpace(ent.Properties[propID])
}

// setEntityState changes an entity's state, tells every player on its grid
// that knows it, and re-evaluates open_when conditions there.
func (s *serverInterface) setEntityState(ent *Entity, state string) {
	s.setEntityStateDepth(ent, state, 0)
}

func (s *serverInterface) setEntityStateDepth(ent *Entity, state string, depth int) {
	if ent == nil || ent.State == state {
		return
	}
	ent.State = state
	change := &pb.EntityStateChange{EntityId: ent.ServerId, NewState: state}
	for _, p := range s.remotePlayers {
		if p.currentFrame == nil || p.CurrentWorldId != ent.WorldId || p.CurrentGridId != ent.GridId {
			continue
		}
		if _, known := p.KnownEntities[ent.ServerId]; known {
			p.currentFrame.EntityStateChanges = append(p.currentFrame.EntityStateChanges, change)
		}
	}
	// Chains (a door opening another) settle quickly; the depth bound only
	// stops badly wired maps from looping forever.
	if depth < 8 {
		s.evaluateOpenWhen(ent.WorldId, ent.GridId, depth+1)
	}
}

// condition is one "id=state" term of open_when.
type condition struct{ ID, State string }

func parseOpenWhen(raw string) ([]condition, error) {
	var out []condition
	for _, term := range strings.Split(raw, ",") {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		id, state, ok := strings.Cut(term, "=")
		if !ok || strings.TrimSpace(id) == "" || strings.TrimSpace(state) == "" {
			return nil, fmt.Errorf("open_when term %q: want id=state", term)
		}
		out = append(out, condition{strings.TrimSpace(id), strings.TrimSpace(state)})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("open_when is empty")
	}
	return out, nil
}

// evaluateOpenWhen opens or closes every entity on the grid that has an
// open_when property, according to the current states of the ids it names.
func (s *serverInterface) evaluateOpenWhen(worldId, gridId string, depth int) {
	ents := s.entitiesOnGrid(worldId, gridId)
	byID := map[string]*Entity{}
	for _, e := range ents {
		if id := entityID(e); id != "" {
			byID[id] = e
		}
	}
	sort.Slice(ents, func(i, j int) bool { return ents[i].ServerId < ents[j].ServerId })
	for _, e := range ents {
		raw := e.Properties[propOpenWhen]
		if raw == "" {
			continue
		}
		conds, err := parseOpenWhen(raw)
		if err != nil {
			log.Warnf("entity %d (%s): %v", e.ServerId, e.Definition, err)
			continue
		}
		open := true
		for _, c := range conds {
			if other := byID[c.ID]; other == nil || other.State != c.State {
				open = false
				break
			}
		}
		want := "closed"
		if open {
			want = "open"
		}
		s.setEntityStateDepth(e, want, depth)
	}
}

// parseLoot reads "item[:count], ..." (count defaults to 1).
func parseLoot(raw string) []entdefs.LootDrop {
	var out []entdefs.LootDrop
	for _, term := range strings.Split(raw, ",") {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		item, countStr, hasCount := strings.Cut(term, ":")
		count := int64(1)
		if hasCount {
			if n, err := strconv.ParseInt(strings.TrimSpace(countStr), 10, 32); err == nil && n > 0 {
				count = n
			}
		}
		out = append(out, entdefs.LootDrop{Item: strings.TrimSpace(item), Count: int32(count)})
	}
	return out
}

// lootFor is the entity's loot: its map `loot` property, else the entdef's.
func lootFor(ent *Entity, entdef *entdefs.EntityDefinition) []entdefs.LootDrop {
	if raw, ok := ent.Properties[propLoot]; ok {
		return parseLoot(raw)
	}
	if entdef == nil {
		return nil
	}
	return entdef.Loot
}

func (s *serverInterface) grantLoot(rp *RemotePlayer, loot []entdefs.LootDrop) string {
	var names []string
	for _, drop := range loot {
		count := drop.Count
		if count <= 0 {
			count = 1
		}
		rp.addInventoryItem(drop.Item, count)
		name := itemTitle(s, drop.Item)
		if count > 1 {
			name = fmt.Sprintf("%s x%d", name, count)
		}
		names = append(names, name)
	}
	if len(loot) > 0 {
		rp.PendingInventoryUpdate = true
	}
	return strings.Join(names, ", ")
}

func sayAt(rp *RemotePlayer, ent *Entity, text string) {
	rp.PendingHudMessage = &text
	rp.PendingHudEntityId = ent.ServerId
}

// interactBehaviour runs the entdef's interact rules (toggle / open).
// Returns whether anything happened.
func (s *serverInterface) interactBehaviour(rp *RemotePlayer, ent *Entity, entdef *entdefs.EntityDefinition) bool {
	in := entdef.Interact
	if in == nil {
		return false
	}
	if len(in.Toggle) > 0 {
		next := in.Toggle[0]
		for i, st := range in.Toggle {
			if st == ent.State {
				next = in.Toggle[(i+1)%len(in.Toggle)]
				break
			}
		}
		s.setEntityState(ent, next)
		return true
	}
	if in.Open != "" {
		if ent.State == in.Open {
			if in.Close == "" {
				return false // opens once: nothing more to do
			}
			s.setEntityState(ent, in.Close)
			return true
		}
		// Reopening a closed-again chest: no key, no loot, no message.
		if in.Close != "" && ent.State == in.Close {
			s.setEntityState(ent, in.Open)
			return true
		}
		if req := strings.TrimSpace(ent.Properties[propRequires]); req != "" {
			if rp.inventoryCount(req) <= 0 {
				msg := ent.Properties[propLockedMsg]
				if msg == "" {
					msg = "It's locked."
				}
				sayAt(rp, ent, msg)
				return true
			}
			rp.removeInventoryItem(req, 1)
			rp.PendingInventoryUpdate = true
		}
		s.setEntityState(ent, in.Open)
		got := s.grantLoot(rp, lootFor(ent, nil))
		msg := ent.Properties[propMsg]
		switch {
		case msg != "" && got != "":
			msg += " (Got " + got + ")"
		case got != "":
			msg = "Got " + got
		}
		if msg != "" {
			sayAt(rp, ent, msg)
		}
		return true
	}
	return false
}
