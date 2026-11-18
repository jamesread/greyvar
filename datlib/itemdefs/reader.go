package itemdefs

import (
	"os"
	"path/filepath"
	"strings"

	datlib "github.com/jamesread/greyvar/datlib/common"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func ItemsDir() string {
	return filepath.Join(datlib.DatDir(), "itemdefs")
}

func ReadItemdef(name string) (*ItemDefinition, error) {
	return ReadItemdefFile(filepath.Join(ItemsDir(), name+".yml"))
}

func ReadItemdefFile(filename string) (*ItemDefinition, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	item := &ItemDefinition{}
	if err := yaml.UnmarshalStrict(file, item); err != nil {
		log.Warnf("Cannot unmarshal itemdef %v: %v", filename, err)
		return nil, err
	}

	if item.ID == "" {
		base := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
		item.ID = base
	}

	if item.StackMax <= 0 {
		item.StackMax = 99
	}

	return item, nil
}

func LoadAllItemdefs() (map[string]*ItemDefinition, error) {
	dir := ItemsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]*ItemDefinition{}, nil
		}
		return nil, err
	}

	out := make(map[string]*ItemDefinition)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yml")
		item, err := ReadItemdef(name)
		if err != nil {
			log.Warnf("Skipping itemdef %q: %v", name, err)
			continue
		}
		out[item.ID] = item
	}

	return out, nil
}
