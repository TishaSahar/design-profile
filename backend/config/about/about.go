// Package about loads designer contact info from a YAML file that is deployed
// alongside the main config. Edit backend/config/about/contacts.yaml to update
// the public contact directory.
package about

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Contacts holds the designer's public contact information.
type Contacts struct {
	Email     string `yaml:"email"`
	Instagram string `yaml:"instagram"`
	Telegram  string `yaml:"telegram"`
}

// LoadContacts reads and parses the YAML file at path.
func LoadContacts(path string) (*Contacts, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read contacts file %q: %w", path, err)
	}
	var c Contacts
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse contacts file: %w", err)
	}
	return &c, nil
}
