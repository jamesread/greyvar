package greyvarserver

import (
	"github.com/jamesread/greyvar/datlib/dialogs"
	"github.com/jamesread/greyvar/datlib/rigs"
	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

// activeDialog is where a player is in a conversation with an NPC.
type activeDialog struct {
	entityId int64
	name     string
	node     string
}

func (s *serverInterface) dialog(name string) *dialogs.Dialog {
	if d, ok := s.dialogs[name]; ok {
		return d
	}
	d, err := dialogs.Read(name)
	if err != nil {
		log.Warnf("dialog %q: %v", name, err)
		return nil
	}
	if s.dialogs == nil {
		s.dialogs = map[string]*dialogs.Dialog{}
	}
	s.dialogs[name] = d
	return d
}

// startDialog opens an NPC's conversation for rp (interact key).
func (s *serverInterface) startDialog(rp *RemotePlayer, ent *Entity, name string) bool {
	d := s.dialog(name)
	if d == nil {
		return false
	}
	s.faceTowards(ent, rp)
	s.sendDialogNode(rp, ent, name, d, d.Start)
	return true
}

func (s *serverInterface) sendDialogNode(rp *RemotePlayer, ent *Entity, name string, d *dialogs.Dialog, nodeName string) {
	node := d.Nodes[nodeName]
	text := dialogs.FillBalances(node.Text, func(item string) int64 {
		return s.accountBalance(rp.Username, item)
	})
	if node.Deposit != "" {
		text += "\n" + s.depositNote(rp, node.Deposit)
	}
	if node.Give != "" {
		if got := s.grantLoot(rp, parseLoot(node.Give)); got != "" {
			text += "\n(Got " + got + ")"
		}
	}
	msg := &pb.Dialog{EntityId: ent.ServerId, Node: nodeName, Speaker: d.Speaker, Text: text}
	for _, o := range node.Options {
		label := o.Label
		if len(o.Wear) > 0 && s.wearing(rp, o.Wear) {
			label += " (wearing)"
		}
		msg.Options = append(msg.Options, &pb.DialogOption{Label: label, Ends: o.Goto == "", Wear: o.Wear})
	}
	rp.dialog = &activeDialog{entityId: ent.ServerId, name: name, node: nodeName}
	rp.currentFrame.Dialog = msg
}

// processDialogChoiceRequest follows the picked option. The choice must be
// for the node this player was last shown, and the NPC still within reach.
func processDialogChoiceRequest(s *serverInterface, rp *RemotePlayer, req *pb.DialogChoiceRequest) {
	at := rp.dialog
	if at == nil || at.entityId != req.GetEntityId() || at.node != req.GetNode() {
		return
	}
	ent := s.entityInstances[at.entityId]
	d := s.dialog(at.name)
	if ent == nil || d == nil || ent.WorldId != rp.CurrentWorldId || ent.GridId != rp.CurrentGridId || !s.inInteractRange(rp.Entity, ent) {
		rp.dialog = nil
		return
	}
	node := d.Nodes[at.node]
	i := int(req.GetOption())
	if node == nil || i < 0 || i >= len(node.Options) {
		return
	}
	if wear := node.Options[i].Wear; len(wear) > 0 {
		s.applyWear(rp, wear)
	}
	next := node.Options[i].Goto
	if next == "" {
		rp.dialog = nil
		return
	}
	s.faceTowards(ent, rp)
	s.sendDialogNode(rp, ent, at.name, d, next)
}

// splitWear sorts a dialog option's wear into rig layer variants (styles)
// and colour slot palettes. "hair" is both a layer and a colour slot, so the
// value decides: a variant of the layer (hair: long) is a style, anything
// else (hair: blonde) a colour.
func (s *serverInterface) splitWear(r *rigs.Rig, wear map[string]string) (parts, palettes map[string]string) {
	parts, palettes = map[string]string{}, map[string]string{}
	for k, v := range wear {
		if isLayerVariant(r, k, v) {
			parts[k] = v
		} else if _, ok := r.Palettes[k]; ok {
			palettes[k] = v
		} else {
			parts[k] = v // unknown: Apply reports it
		}
	}
	return parts, palettes
}

func isLayerVariant(r *rigs.Rig, layer, variant string) bool {
	for _, l := range r.Layers {
		if l.Name != layer {
			continue
		}
		for _, v := range l.Variants {
			if v == variant {
				return true
			}
		}
	}
	return false
}

func (s *serverInterface) currentLook(rp *RemotePlayer) rigs.Appearance {
	if rp.Entity != nil && rp.Entity.Appearance != nil {
		return *rp.Entity.Appearance
	}
	return s.defaultLook(rp.Username)
}

// applyWear changes rp's look (e.g. a tailor's new shirt) for everyone.
func (s *serverInterface) applyWear(rp *RemotePlayer, wear map[string]string) {
	if s.rigs == nil || rp.Entity == nil {
		return
	}
	look := s.currentLook(rp)
	r := s.rigFor(&look)
	parts, palettes := s.splitWear(r, wear)
	next, err := r.Apply(look, parts, palettes)
	if err != nil {
		log.Warnf("dialog wear %v: %v", wear, err)
		return
	}
	setAppearance(s, rp, next)
}

// wearing reports whether rp already looks like wear.
func (s *serverInterface) wearing(rp *RemotePlayer, wear map[string]string) bool {
	if s.rigs == nil {
		return false
	}
	look := s.currentLook(rp)
	parts, palettes := s.splitWear(s.rigFor(&look), wear)
	for k, v := range parts {
		if look.Parts[k] != v {
			return false
		}
	}
	for k, v := range palettes {
		if look.Palettes[k] != v {
			return false
		}
	}
	return true
}
