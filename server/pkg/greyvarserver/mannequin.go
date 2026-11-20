package greyvarserver

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

// Mannequins (or any figure placed with a `wear` map property) show a real
// outfit, and using one puts that outfit on the player. The property is the
// same as a dialog option's wear, written as key=value pairs:
//
//	wear: outfit=dress,garment=blue,trim=white
//	title: blue dress          (optional, for the message)

const propWear = "wear"

// parseWear reads "k=v,k=v" (spaces allowed); nil when empty.
func parseWear(s string) map[string]string {
	var out map[string]string
	for _, kv := range strings.Split(s, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(kv), "=")
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if !ok || k == "" || v == "" {
			continue
		}
		if out == nil {
			out = map[string]string{}
		}
		out[k] = v
	}
	return out
}

// entityWear is the outfit an entity shows (its `wear` property).
func entityWear(ent *Entity) map[string]string {
	if ent == nil {
		return nil
	}
	return parseWear(ent.Properties[propWear])
}

// dressFigure puts an entity's `wear` on its own look (a mannequin in the
// outfit it sells).
func (s *serverInterface) dressFigure(ent *Entity) {
	wear := entityWear(ent)
	if len(wear) == 0 || ent.Appearance == nil || s.rigs == nil {
		return
	}
	r := s.rigs.For(ent.Appearance)
	parts, palettes := s.splitWear(r, wear)
	a, err := r.Apply(*ent.Appearance, parts, palettes)
	if err != nil {
		log.Warnf("%s at (%d,%d) on %s: wear %v: %v", ent.Definition, ent.X/16, ent.Y/16, ent.GridId, wear, err)
		return
	}
	ent.Appearance = &a
}

// tryOn puts the figure's outfit on the player (the use key on a mannequin).
func (s *serverInterface) tryOn(rp *RemotePlayer, ent *Entity, wear map[string]string) bool {
	title := strings.TrimSpace(ent.Properties["title"])
	if title == "" {
		title = "outfit"
	}
	var text string
	if s.wearing(rp, wear) {
		text = "You're already wearing the " + title + "."
	} else {
		s.applyWear(rp, wear)
		text = "You put on the " + title + "."
	}
	rp.PendingHudMessage = &text
	rp.PendingHudEntityId = rp.Entity.ServerId
	return true
}
