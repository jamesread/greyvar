// Package dialogs loads NPC conversations: server/dat/dialogs/<name>.yml.
//
//	speaker: Bill
//	start: hello
//	nodes:
//	  hello:
//	    text: "Morning!"
//	    options:
//	      - {label: "Got any seeds?", goto: seeds}
//	      - {label: "Bye"}            # no goto: ends the conversation
//	  seeds:
//	    text: "Here you go."
//	    give: "hay_seeds:5"            # item[:count], ... granted on arrival
//	    options:
//	      - {label: "Thanks!"}
//	      - {label: "Blue, please", wear: {shirt: blue}, goto: done}
//
// An option's wear changes the player's look when picked: keys are colour
// slots (shirt: blue) or layers (hair: long), values as in rig.yml.
//
// A node's deposit moves every one of that item the player carries into
// their account (kept by the server across sessions), e.g. a banker:
//
//	deposit:
//	  text: "Let me count those."
//	  deposit: rift_coin
//
// Node text may show an account balance with {balance:<item>}.
package dialogs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	datlib "github.com/jamesread/greyvar/datlib/common"
	"gopkg.in/yaml.v3"
)

type Dialog struct {
	Speaker string           `yaml:"speaker"`
	Start   string           `yaml:"start"`
	Nodes   map[string]*Node `yaml:"nodes"`
}

type Node struct {
	Text    string   `yaml:"text"`
	Give    string   `yaml:"give,omitempty"`
	Deposit string   `yaml:"deposit,omitempty"` // item id: all carried go into the account
	Options []Option `yaml:"options"`
}

type Option struct {
	Label string            `yaml:"label"`
	Goto  string            `yaml:"goto,omitempty"`
	Wear  map[string]string `yaml:"wear,omitempty"`
}

// Read loads dat/dialogs/<name>.yml and validates it.
func Read(name string) (*Dialog, error) {
	return ReadFile(filepath.Join(datlib.DatDir(), "dialogs", name+".yml"))
}

func ReadFile(path string) (*Dialog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d Dialog
	if err := yaml.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if errs := d.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("%s: %v", path, errs)
	}
	return &d, nil
}

// Validate checks the start node, every goto, and that nodes have options.
func (d *Dialog) Validate() []string {
	var errs []string
	if d.Nodes[d.Start] == nil {
		errs = append(errs, fmt.Sprintf("start node %q does not exist", d.Start))
	}
	names := make([]string, 0, len(d.Nodes))
	for n := range d.Nodes {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		node := d.Nodes[n]
		if node == nil || len(node.Options) == 0 {
			errs = append(errs, fmt.Sprintf("node %q needs at least one option", n))
			continue
		}
		for _, o := range node.Options {
			if o.Label == "" {
				errs = append(errs, fmt.Sprintf("node %q: option without a label", n))
			}
			for k, v := range o.Wear {
				if k == "" || v == "" {
					errs = append(errs, fmt.Sprintf("node %q: option %q has an empty wear entry", n, o.Label))
				}
			}
			if o.Goto != "" && d.Nodes[o.Goto] == nil {
				errs = append(errs, fmt.Sprintf("node %q: option %q goes to unknown node %q", n, o.Label, o.Goto))
			}
		}
	}
	return errs
}

var reBalance = regexp.MustCompile(`\{balance:([a-z0-9_]+)\}`)

// BalanceItems lists the items whose balance a text shows ({balance:item}).
func BalanceItems(text string) []string {
	var out []string
	for _, m := range reBalance.FindAllStringSubmatch(text, -1) {
		out = append(out, m[1])
	}
	return out
}

// FillBalances replaces each {balance:item} in text with balance(item).
func FillBalances(text string, balance func(item string) int64) string {
	return reBalance.ReplaceAllStringFunc(text, func(m string) string {
		return strconv.FormatInt(balance(reBalance.FindStringSubmatch(m)[1]), 10)
	})
}

// Items are the item ids a dialog deposits or shows a balance of.
func (d *Dialog) Items() []string {
	seen := map[string]bool{}
	var out []string
	for _, n := range d.Nodes {
		items := BalanceItems(n.Text)
		if n.Deposit != "" {
			items = append(items, n.Deposit)
		}
		for _, it := range items {
			if !seen[it] {
				seen[it] = true
				out = append(out, it)
			}
		}
	}
	sort.Strings(out)
	return out
}
