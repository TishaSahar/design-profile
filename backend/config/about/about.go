// Package about loads designer content from files deployed alongside the main
// config. Edit backend/config/about/contacts.yaml for contact info,
// backend/config/about/bio.txt for the about-page text, and place the profile
// photo at the path referenced by the "photo" field in contacts.yaml.
package about

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Contacts holds the designer's public contact information and photo filename.
type Contacts struct {
	Email     string `yaml:"email"`
	Instagram string `yaml:"instagram"`
	Telegram  string `yaml:"telegram"`
	Photo     string `yaml:"photo"`
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

// LoadBio reads the bio text file at path and trims surrounding whitespace.
// Returns an empty string without error if the file does not exist.
func LoadBio(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read bio file %q: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}
