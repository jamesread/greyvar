package greyvarserver

import (
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/jamesread/greyvar/server/gen/greyvarprotocol"
)

func tailorServer(t *testing.T) *serverInterface {
	t.Helper()
	dat, _ := filepath.Abs(filepath.Join("..", "..", "dat"))
	res, _ := filepath.Abs(filepath.Join("..", "..", "..", "res"))
	t.Setenv("GREYVAR_DAT_DIR", dat)
	t.Setenv("GREYVAR_RES_DIR", res)
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	t.Setenv("GREYVAR_WORLD", "isleOfStarting_dev_tiled")
	saved := playerHitbox // newServer sets it from the rig; other tests expect the default
	t.Cleanup(func() { playerHitbox = saved })
	return newServer()
}

// The tailor's shop on the 0.0 hill: in through its door, out through the
// doorway, and the house is solid.
func TestTailorShopDoorRoundTrip(t *testing.T) {
	s := tailorServer(t)
	const world = "isleOfStarting_dev_tiled"
	player := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: 7 * 16, Y: 12 * 16, WorldId: world, GridId: "0.0.tmj"}
	s.entityInstances[player.ServerId] = player
	rp := &RemotePlayer{Username: "guest", Entity: player, CurrentWorldId: world, CurrentGridId: "0.0.tmj",
		KnownEntities: map[int64]*Entity{}, KnownEntdefs: map[string]bool{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"guest": rp}
	for _, col := range []int32{6, 8, 9} {
		if s.isTraversableAt(rp, col*16, 11*16) {
			t.Errorf("shop wall at (11,%d) should block", col)
		}
	}
	walk := func(dy int32) {
		from := rp.CurrentGridId
		for i := 0; i < 12 && rp.CurrentGridId == from; i++ {
			s.frameTime += 70_000_000
			processMoveRequest(s, rp, &pb.MoveRequest{Y: dy})
		}
	}
	walk(-1)
	if rp.CurrentGridId != "0.0.tailorShop.tmj" || player.X != 6*16 || player.Y != 10*16 {
		t.Fatalf("into the shop at (10,6), got %s (%d,%d)", rp.CurrentGridId, player.Y/16, player.X/16)
	}
	if rp.PendingGridTransition != nil {
		t.Fatal("a door fades, it does not scroll")
	}
	walk(+1)
	if rp.CurrentGridId != "0.0.tmj" || player.X != 7*16 || player.Y != 12*16 {
		t.Fatalf("back out in front of the door (12,7), got %s (%d,%d)", rp.CurrentGridId, player.Y/16, player.X/16)
	}
}

// Geraldine the tailor changes your clothes.
func TestGeraldineChangesClothes(t *testing.T) {
	s := tailorServer(t)
	const world, grid = "isleOfStarting_dev_tiled", "0.0.tailorShop.tmj"
	var geraldine *Entity
	for _, e := range s.entitiesOnGrid(world, grid) {
		if e.Definition == "geraldine" {
			geraldine = e
		}
	}
	if geraldine == nil {
		t.Fatal("Geraldine is not in the tailor's shop")
	}
	if a := geraldine.Appearance; a == nil || a.Parts["hair"] != "bun" || a.Palettes["hair"] != "red" {
		t.Fatalf("Geraldine's fixed look: %+v", a)
	}
	rp := playerOn(s, grid, geraldine.X+16, geraldine.Y)
	look := s.defaultLook("visitor")
	look.Palettes["shirt"] = "magenta"
	rp.Entity.Appearance = &look
	choose := func(node string, option int32) *pb.Dialog {
		rp.currentFrame = &pb.ServerUpdate{}
		processDialogChoiceRequest(s, rp, &pb.DialogChoiceRequest{EntityId: geraldine.ServerId, Node: node, Option: option})
		return rp.currentFrame.Dialog
	}

	if !processEntityInteract(s, rp, geraldine.ServerId) || rp.currentFrame.Dialog.Speaker != "Geraldine" {
		t.Fatal("talking to Geraldine should open her dialog")
	}
	shirts := choose("hello", 1)
	if shirts == nil || shirts.Node != "shirts" || !strings.HasSuffix(shirts.Options[0].Label, "(wearing)") || strings.HasSuffix(shirts.Options[1].Label, "(wearing)") {
		t.Fatalf("shirt choices should mark the current magenta: %+v", shirts)
	}
	done := choose("shirts", 1) // Blue
	if done == nil || done.Node != "done" || rp.Entity.Appearance.Palettes["shirt"] != "blue" {
		t.Fatalf("picking Blue should change the shirt: dialog %+v, look %+v", done, rp.Entity.Appearance)
	}
	if rp.Entity.Appearance.Palettes["pants"] != look.Palettes["pants"] {
		t.Fatal("only the shirt should change")
	}
	if len(rp.PendingAppearanceChanges) == 0 {
		t.Fatal("the new look should be sent to everyone who can see the player")
	}
	again := choose("done", 2) // Another shirt
	if !strings.HasSuffix(again.Options[1].Label, "(wearing)") {
		t.Fatalf("Blue is now worn: %+v", again.Options)
	}
}

// Professor Ivan's house beside the tailor's shop: in and out through its
// door, and he changes your skin tone.
func TestProfessorIvanChangesSkinTone(t *testing.T) {
	s := tailorServer(t)
	const world = "isleOfStarting_dev_tiled"
	player := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: 14 * 16, Y: 11 * 16, WorldId: world, GridId: "0.0.tmj"}
	s.entityInstances[player.ServerId] = player
	rp := &RemotePlayer{Username: "guest", Entity: player, CurrentWorldId: world, CurrentGridId: "0.0.tmj",
		KnownEntities: map[int64]*Entity{}, KnownEntdefs: map[string]bool{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"guest": rp}
	walk := func(dy int32) {
		from := rp.CurrentGridId
		for i := 0; i < 12 && rp.CurrentGridId == from; i++ {
			s.frameTime += 70_000_000
			processMoveRequest(s, rp, &pb.MoveRequest{Y: dy})
		}
	}
	walk(-1)
	if rp.CurrentGridId != "0.0.professorHouse.tmj" || player.X != 6*16 || player.Y != 10*16 {
		t.Fatalf("into Ivan's house at (10,6), got %s (%d,%d)", rp.CurrentGridId, player.Y/16, player.X/16)
	}

	var ivan *Entity
	for _, e := range s.entitiesOnGrid(world, "0.0.professorHouse.tmj") {
		if e.Definition == "ivan" {
			ivan = e
		}
	}
	if ivan == nil || ivan.Appearance == nil || ivan.Appearance.Palettes["skin"] != "brown" {
		t.Fatalf("Professor Ivan with his fixed look should be inside, got %+v", ivan)
	}
	look := s.defaultLook("guest")
	look.Palettes["skin"] = "rose"
	player.Appearance = &look
	player.X, player.Y = ivan.X+16, ivan.Y
	if !processEntityInteract(s, rp, ivan.ServerId) || rp.currentFrame.Dialog.Speaker != "Professor Ivan" {
		t.Fatal("talking to Professor Ivan should open his dialog")
	}
	choose := func(node string, option int32) *pb.Dialog {
		rp.currentFrame = &pb.ServerUpdate{}
		processDialogChoiceRequest(s, rp, &pb.DialogChoiceRequest{EntityId: ivan.ServerId, Node: node, Option: option})
		return rp.currentFrame.Dialog
	}
	tones := choose("hello", 0)
	if tones == nil || tones.Node != "tones" || !strings.HasSuffix(tones.Options[0].Label, "(wearing)") {
		t.Fatalf("tone choices should mark the current rose: %+v", tones)
	}
	if choose("tones", 4); player.Appearance.Palettes["skin"] != "deep" || player.Appearance.Palettes["shirt"] != look.Palettes["shirt"] {
		t.Fatalf("picking Deep should change only the skin tone: %+v", player.Appearance)
	}

	player.X, player.Y = 6*16, 10*16
	walk(+1)
	if rp.CurrentGridId != "0.0.tmj" || player.X != 14*16 || player.Y != 11*16 {
		t.Fatalf("back out in front of his door (11,14), got %s (%d,%d)", rp.CurrentGridId, player.Y/16, player.X/16)
	}
}

// Both shop doors on the 0.0 hill can be walked to from where the hill
// enters 0.0 (its bottom edge, up from 1.0's steps).
func TestHillHouseDoorsAreReachable(t *testing.T) {
	s := tailorServer(t)
	const world, grid = "isleOfStarting_dev_tiled", "0.0.tmj"
	rp := &RemotePlayer{CurrentWorldId: world, CurrentGridId: grid, KnownEntities: map[int64]*Entity{}, currentFrame: &pb.ServerUpdate{}}
	g := s.gridById(world, grid)
	type pos struct{ x, y int32 }
	seen := map[pos]bool{}
	var q []pos
	for c := int32(0); c < 16; c++ {
		if tl := baseTile(g, 15, uint32(c)); tl != nil && tl.FrameIndex == 70 && s.isTraversableAt(rp, c*16, 15*16) {
			seen[pos{c * 16, 15 * 16}] = true
			q = append(q, pos{c * 16, 15 * 16})
		}
	}
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		for _, d := range []pos{{2, 0}, {-2, 0}, {0, 2}, {0, -2}} {
			n := pos{cur.x + d.x, cur.y + d.y}
			if !seen[n] && s.isTraversableAt(rp, n.x, n.y) {
				seen[n] = true
				q = append(q, n)
			}
		}
	}
	for name, at := range map[string]pos{"Geraldine's door": {7 * 16, 12 * 16}, "Professor Ivan's door": {14 * 16, 11 * 16}} {
		if !seen[at] {
			t.Errorf("%s: the spot in front of it (%d,%d) cannot be reached", name, at.y/16, at.x/16)
		}
	}
}

// Claire the hairdresser (salon on 1.2): style and colour both use the
// "hair" key; the value decides which.
func TestClaireChangesHairstyleAndColour(t *testing.T) {
	s := tailorServer(t)
	const world = "isleOfStarting_dev_tiled"
	player := &Entity{ServerId: s.nextEntityId(), Definition: "player", X: 12 * 16, Y: 6 * 16, WorldId: world, GridId: "1.2.tmj"}
	s.entityInstances[player.ServerId] = player
	rp := &RemotePlayer{Username: "guest", Entity: player, CurrentWorldId: world, CurrentGridId: "1.2.tmj",
		KnownEntities: map[int64]*Entity{}, KnownEntdefs: map[string]bool{}, currentFrame: &pb.ServerUpdate{}}
	s.remotePlayers = map[string]*RemotePlayer{"guest": rp}
	walk := func(dy int32) {
		from := rp.CurrentGridId
		for i := 0; i < 12 && rp.CurrentGridId == from; i++ {
			s.frameTime += 70_000_000
			processMoveRequest(s, rp, &pb.MoveRequest{Y: dy})
		}
	}
	walk(-1)
	if rp.CurrentGridId != "1.2.salon.tmj" || player.X != 6*16 || player.Y != 10*16 {
		t.Fatalf("into the salon at (10,6), got %s (%d,%d)", rp.CurrentGridId, player.Y/16, player.X/16)
	}
	var claire *Entity
	for _, e := range s.entitiesOnGrid(world, "1.2.salon.tmj") {
		if e.Definition == "claire" {
			claire = e
		}
	}
	if claire == nil || claire.Appearance == nil || claire.Appearance.Palettes["hair"] != "blonde" || claire.Appearance.Parts["hair"] != "bob" {
		t.Fatalf("Claire with her fixed look should be in the salon, got %+v", claire)
	}
	look := s.defaultLook("guest")
	look.Parts["hair"], look.Palettes["hair"] = "short", "brown"
	player.Appearance = &look
	player.X, player.Y = claire.X+16, claire.Y
	if !processEntityInteract(s, rp, claire.ServerId) || rp.currentFrame.Dialog.Speaker != "Claire" {
		t.Fatal("talking to Claire should open her dialog")
	}
	choose := func(node string, option int32) *pb.Dialog {
		rp.currentFrame = &pb.ServerUpdate{}
		processDialogChoiceRequest(s, rp, &pb.DialogChoiceRequest{EntityId: claire.ServerId, Node: node, Option: option})
		return rp.currentFrame.Dialog
	}
	styles := choose("hello", 0)
	if !strings.HasSuffix(styles.Options[0].Label, "(wearing)") {
		t.Fatalf("short is the current style: %+v", styles.Options)
	}
	choose("styles", 1) // Long
	if player.Appearance.Parts["hair"] != "long" || player.Appearance.Palettes["hair"] != "brown" {
		t.Fatalf("Long should change the style only: %+v", player.Appearance)
	}
	choose("done", 1)    // Another colour
	choose("colours", 3) // Red
	if player.Appearance.Palettes["hair"] != "red" || player.Appearance.Parts["hair"] != "long" {
		t.Fatalf("Red should change the colour only: %+v", player.Appearance)
	}
	// Every adult style, then a beard (in the hair colour) and off again.
	styles = choose("done", 0) // Another style
	var labels []string
	for _, o := range styles.Options {
		labels = append(labels, o.Label)
	}
	if got := strings.Join(labels, ","); !strings.Contains(got, "Ponytail") || !strings.Contains(got, "Bun") ||
		!strings.Contains(got, "Curly") || !strings.Contains(got, "Bob") || !strings.Contains(got, "Shave it all off") {
		t.Fatalf("Claire should offer every adult style: %s", got)
	}
	choose("styles", 6) // Shave it all off
	if player.Appearance.Parts["hair"] != "bald" {
		t.Fatalf("bald: %+v", player.Appearance)
	}
	choose("done", 2) // A beard
	choose("beard", 0)
	if player.Appearance.Parts["beard"] != "beard" || player.Appearance.Palettes["hair"] != "red" {
		t.Fatalf("Grow a beard should add a (red) beard: %+v", player.Appearance)
	}
	choose("done", 2)
	choose("beard", 1)
	if player.Appearance.Parts["beard"] != "none" {
		t.Fatalf("Clean shaven should remove it: %+v", player.Appearance)
	}

	player.X, player.Y = 6*16, 10*16
	walk(+1)
	if rp.CurrentGridId != "1.2.tmj" || player.X != 12*16 || player.Y != 6*16 {
		t.Fatalf("back out in front of the salon (6,12), got %s (%d,%d)", rp.CurrentGridId, player.Y/16, player.X/16)
	}
}

// Geraldine sells dresses: the outfit layer (drawn over shirt and trousers)
// in a garment colour with a trim colour. A shirt, or "Back to a shirt and
// trousers", takes the dress off again; the shirt and trousers underneath
// keep their colours.
func TestGeraldineSellsDresses(t *testing.T) {
	s := tailorServer(t)
	const world, grid = "isleOfStarting_dev_tiled", "0.0.tailorShop.tmj"
	var geraldine *Entity
	for _, e := range s.entitiesOnGrid(world, grid) {
		if e.Definition == "geraldine" {
			geraldine = e
		}
	}
	if a := geraldine.Appearance; a == nil || a.Parts["outfit"] != "dress" || a.Palettes["garment"] != "purple" {
		t.Fatalf("Geraldine wears one of her dresses: %+v", a)
	}
	rp := playerOn(s, grid, geraldine.X+16, geraldine.Y)
	look := s.defaultLook("visitor")
	if look.Parts["outfit"] != "none" {
		t.Fatalf("new players start without an outfit: %+v", look)
	}
	look.Palettes["shirt"], look.Palettes["pants"] = "blue", "denim"
	rp.Entity.Appearance = &look
	choose := func(node string, option int32) *pb.Dialog {
		rp.currentFrame = &pb.ServerUpdate{}
		processDialogChoiceRequest(s, rp, &pb.DialogChoiceRequest{EntityId: geraldine.ServerId, Node: node, Option: option})
		return rp.currentFrame.Dialog
	}
	processEntityInteract(s, rp, geraldine.ServerId)
	if d := choose("hello", 0); d == nil || d.Node != "dresses" {
		t.Fatalf("A dress: %+v", d)
	}
	if d := choose("dresses", 4); d == nil || d.Node != "trims" { // Pink
		t.Fatalf("a colour leads to the trim: %+v", d)
	}
	choose("trims", 0) // Gold
	a := rp.Entity.Appearance
	if a.Parts["outfit"] != "dress" || a.Palettes["garment"] != "pink" || a.Palettes["trim"] != "gold" {
		t.Fatalf("a pink dress with gold trim: %+v", a)
	}
	if a.Palettes["shirt"] != "blue" || a.Palettes["pants"] != "denim" {
		t.Fatalf("the shirt and trousers underneath keep their colours: %+v", a)
	}
	choose("done", 1) // Back to a shirt and trousers
	if a := rp.Entity.Appearance; a.Parts["outfit"] != "none" || a.Palettes["shirt"] != "blue" {
		t.Fatalf("taking the dress off shows the shirt again: %+v", a)
	}
	choose("done", 0) // Another dress
	choose("dresses", 0)
	choose("trims", 1)
	choose("done", 2)   // Another shirt
	choose("shirts", 2) // Green
	if a := rp.Entity.Appearance; a.Parts["outfit"] != "none" || a.Palettes["shirt"] != "green" {
		t.Fatalf("buying a shirt takes the dress off: %+v", a)
	}
}

// Each shop's hanging sign names it while you stand at its door (sticky, like
// a signpost), and walking away clears it.
func TestShopSignsNameTheShops(t *testing.T) {
	s := tailorServer(t)
	const world = "isleOfStarting_dev_tiled"
	cases := []struct {
		grid     string
		row, col int32 // the spot in front of the door
		name     string
	}{
		{"1.2.tmj", 6, 12, "Claire's Hair Salon"},
		{"1.2.tmj", 6, 4, "Bill's Farmhouse"},
		{"0.0.tmj", 12, 7, "Geraldine's Tailor Shop"},
		{"0.0.tmj", 11, 14, "Professor Ivan's Laboratory"},
	}
	for _, c := range cases {
		rp := playerOn(s, c.grid, c.col*16, c.row*16)
		rp.ActiveEntityIds = map[int64]bool{}
		s.processEntityInteractions(rp)
		if rp.PendingHudMessage == nil || *rp.PendingHudMessage != c.name || rp.pendingHudSticky != rp.PendingHudMessage {
			t.Errorf("%s at (%d,%d): want sticky %q, got %v", c.grid, c.row, c.col, c.name, rp.PendingHudMessage)
			continue
		}
		rp.PendingHudMessage = nil
		rp.Entity.Y += 5 * 16 // walk away
		s.processEntityInteractions(rp)
		if rp.PendingHudMessage == nil || *rp.PendingHudMessage != "" {
			t.Errorf("%s: walking away should clear the label, got %v", c.name, rp.PendingHudMessage)
		}
	}
}
