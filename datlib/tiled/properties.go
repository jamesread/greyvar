package tiled

import "strings"

// boolProperty reads a bool tile property (false when absent).
func boolProperty(props []propertyEntry, name string) bool {
	for _, prop := range props {
		if strings.EqualFold(strings.TrimSpace(prop.Name), name) {
			return propertyBoolValue(prop.Type, prop.Value, false)
		}
	}
	return false
}

// decorFromProperties reports a decor=true tile property: decoration that
// sits on top of terrain (flowers, ...) and never collides or hides the
// collision of the terrain beneath it.
func decorFromProperties(props []propertyEntry) bool {
	for _, prop := range props {
		if strings.EqualFold(strings.TrimSpace(prop.Name), "decor") {
			return propertyBoolValue(prop.Type, prop.Value, false)
		}
	}
	return false
}

func drawFromProperties(props []propertyEntry) bool {
	for _, prop := range props {
		if !strings.EqualFold(strings.TrimSpace(prop.Name), "draw") {
			continue
		}
		return propertyBoolValue(prop.Type, prop.Value, true)
	}

	return true
}

func propertyBoolValue(propType string, value interface{}, defaultVal bool) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		default:
			return defaultVal
		}
	case float64:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	default:
		_ = propType
		return defaultVal
	}
}

func tsxPropsToEntries(props []tsxProperty) []propertyEntry {
	if len(props) == 0 {
		return nil
	}

	out := make([]propertyEntry, 0, len(props))
	for _, prop := range props {
		out = append(out, propertyEntry{
			Name:  prop.Name,
			Type:  prop.Type,
			Value: prop.Value,
		})
	}
	return out
}
