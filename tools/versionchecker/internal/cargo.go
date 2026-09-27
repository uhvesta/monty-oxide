package versionchecker

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type cargoManifest struct {
	Workspace struct {
		Members []string `toml:"members"`
		Package struct {
			RustVersion string `toml:"rust-version"`
		} `toml:"package"`
		Dependencies map[string]any `toml:"dependencies"`
	} `toml:"workspace"`
}

func readCargo(root string) (string, []byte, cargoManifest, error) {
	path := filepath.Join(root, "Cargo.toml")
	data, err := os.ReadFile(path)
	var manifest cargoManifest
	if err != nil {
		return "", nil, manifest, err
	}
	if _, err := toml.Decode(string(data), &manifest); err != nil {
		return "", nil, manifest, fmt.Errorf("%s: %w", path, err)
	}
	return path, data, manifest, nil
}

// replaceCargoVersion edits one literal after the TOML parser has identified its pin.
func replaceCargoVersion(data []byte, section, key, field, current, latest string) ([]byte, error) {
	lines := bytes.SplitAfter(data, []byte("\n"))
	active := ""
	offset := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(string(line))
		if strings.HasPrefix(trimmed, "[") {
			if end := strings.IndexByte(trimmed, ']'); end > 0 {
				active = strings.TrimSpace(trimmed[1:end])
			}
		}
		if active != section {
			offset += len(line)
			continue
		}
		equals := bytes.IndexByte(line, '=')
		if equals < 0 || strings.Trim(strings.TrimSpace(string(line[:equals])), `"'`) != key {
			offset += len(line)
			continue
		}
		value := line[equals+1:]
		start, end, ok := tomlStringSpan(value, field)
		if !ok {
			return nil, fmt.Errorf("%s.%s has no editable version string", section, key)
		}
		literal := value[start:end]
		decoded := string(literal[1 : len(literal)-1])
		if literal[0] == '"' {
			var err error
			decoded, err = strconv.Unquote(string(literal))
			if err != nil {
				return nil, err
			}
		}
		if decoded != current {
			return nil, fmt.Errorf("%s.%s pin changed since check", section, key)
		}
		start += offset + equals + 1
		end += offset + equals + 1
		updated := make([]byte, 0, len(data)+len(latest)-len(current))
		updated = append(updated, data[:start]...)
		updated = append(updated, strconv.Quote(latest)...)
		updated = append(updated, data[end:]...)
		return updated, nil
	}
	return nil, fmt.Errorf("%s.%s pin changed since check", section, key)
}

func tomlStringSpan(value []byte, field string) (int, int, bool) {
	if field == "" {
		return quotedSpan(value, 0)
	}
	for i := 0; i < len(value); {
		if value[i] == '#' {
			break
		}
		if value[i] == '"' || value[i] == '\'' {
			_, end, ok := quotedSpan(value, i)
			if !ok {
				return 0, 0, false
			}
			i = end
			continue
		}
		if value[i] >= 'a' && value[i] <= 'z' {
			start := i
			for i < len(value) && ((value[i] >= 'a' && value[i] <= 'z') || value[i] == '-') {
				i++
			}
			if string(value[start:i]) == field {
				for i < len(value) && (value[i] == ' ' || value[i] == '\t') {
					i++
				}
				if i < len(value) && value[i] == '=' {
					return quotedSpan(value, i+1)
				}
			}
			continue
		}
		i++
	}
	return 0, 0, false
}

func quotedSpan(value []byte, from int) (int, int, bool) {
	for from < len(value) && (value[from] == ' ' || value[from] == '\t') {
		from++
	}
	if from >= len(value) || (value[from] != '"' && value[from] != '\'') {
		return 0, 0, false
	}
	quote := value[from]
	for i := from + 1; i < len(value); i++ {
		if quote == '"' && value[i] == '\\' {
			i++
			continue
		}
		if value[i] == quote {
			return from, i + 1, true
		}
	}
	return 0, 0, false
}
