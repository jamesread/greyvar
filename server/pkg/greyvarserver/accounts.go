package greyvarserver

// Accounts: items a player has banked with an NPC (Herman takes rift coins).
// Unlike the inventory, they persist: $GREYVAR_STATE_DIR/accounts.json maps
// username -> item -> count. A dialog node's `deposit: <item>` moves every
// one the player carries into the account; node text shows a balance with
// {balance:<item>} (see datlib/dialogs).

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func accountsPath() string { return filepath.Join(stateDir(), "accounts.json") }

func (s *serverInterface) accountBook() map[string]map[string]int64 {
	if s.accounts == nil {
		s.accounts = map[string]map[string]int64{}
		if data, err := os.ReadFile(accountsPath()); err == nil {
			if err := json.Unmarshal(data, &s.accounts); err != nil {
				log.Warnf("Cannot read %s: %v", accountsPath(), err)
			}
		}
	}
	return s.accounts
}

func (s *serverInterface) saveAccounts() {
	data, err := json.MarshalIndent(s.accountBook(), "", "  ")
	if err == nil {
		err = os.MkdirAll(stateDir(), 0o755)
	}
	if err == nil {
		tmp := accountsPath() + ".tmp"
		if err = os.WriteFile(tmp, data, 0o644); err == nil {
			err = os.Rename(tmp, accountsPath())
		}
	}
	if err != nil {
		log.Warnf("Cannot save accounts: %v", err)
	}
}

// accountBalance is how many of item the player has banked.
func (s *serverInterface) accountBalance(username, item string) int64 {
	return s.accountBook()[username][item]
}

// deposit moves every item the player carries into their account and
// returns how many moved and the new balance.
func (s *serverInterface) deposit(rp *RemotePlayer, item string) (int32, int64) {
	n := rp.inventoryCount(item)
	if n > 0 && rp.removeInventoryItem(item, n) {
		book := s.accountBook()
		if book[rp.Username] == nil {
			book[rp.Username] = map[string]int64{}
		}
		book[rp.Username][item] += int64(n)
		rp.PendingInventoryUpdate = true
		s.saveAccounts()
	} else {
		n = 0
	}
	return n, s.accountBalance(rp.Username, item)
}

// depositNote is the line a deposit adds to the NPC's text.
func (s *serverInterface) depositNote(rp *RemotePlayer, item string) string {
	n, balance := s.deposit(rp, item)
	name := itemTitle(s, item)
	if n == 0 {
		return fmt.Sprintf("(Nothing to deposit. %s balance: %d)", name, balance)
	}
	return fmt.Sprintf("(Deposited %s x%d. Balance: %d)", name, n, balance)
}
