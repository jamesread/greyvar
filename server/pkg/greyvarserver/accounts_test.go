package greyvarserver

import (
	"strings"
	"testing"

	"github.com/jamesread/greyvar/datlib/dialogs"
	"github.com/jamesread/greyvar/datlib/entdefs"
)

func TestDepositBanksRiftCoinsAcrossRestarts(t *testing.T) {
	t.Setenv("GREYVAR_STATE_DIR", t.TempDir())
	s := &serverInterface{}
	rp := &RemotePlayer{Username: "ada"}
	rp.addInventoryItem("rift_coin", 5)
	rp.addInventoryItem("shard", 1)

	note := s.depositNote(rp, "rift_coin")
	if !strings.Contains(note, "x5") || !strings.Contains(note, "Balance: 5") {
		t.Fatalf("note %q", note)
	}
	if rp.inventoryCount("rift_coin") != 0 || rp.inventoryCount("shard") != 1 {
		t.Fatal("only the rift coins should leave the inventory")
	}
	rp.addInventoryItem("rift_coin", 2)
	s.depositNote(rp, "rift_coin")
	if note := s.depositNote(rp, "rift_coin"); !strings.HasPrefix(note, "(Nothing to deposit") || !strings.Contains(note, "7") {
		t.Fatalf("empty deposit note %q", note)
	}

	restarted := &serverInterface{}
	if got := restarted.accountBalance("ada", "rift_coin"); got != 7 {
		t.Fatalf("balance after restart = %d, want 7", got)
	}
	if got := dialogs.FillBalances("You have {balance:rift_coin} coins.", func(item string) int64 {
		return restarted.accountBalance("ada", item)
	}); got != "You have 7 coins." {
		t.Fatalf("balance text %q", got)
	}
	if restarted.accountBalance("bob", "rift_coin") != 0 {
		t.Fatal("accounts are per player")
	}
}

func TestInteractRangeFromEntdef(t *testing.T) {
	s := &serverInterface{entityDefinitions: map[string]*entdefs.EntityDefinition{
		"herman": {Title: "herman", InteractRange: 40},
		"bill":   {Title: "bill"},
	}}
	player := &Entity{X: 0, Y: 32}
	if !s.inInteractRange(player, &Entity{Definition: "herman"}) {
		t.Error("Herman should be reachable across a counter (32px)")
	}
	if s.inInteractRange(player, &Entity{Definition: "bill"}) {
		t.Error("others keep the default 24px")
	}
}
