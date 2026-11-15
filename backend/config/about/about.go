// Package about loads designer content from files deployed alongside the main
// config. Edit backend/config/about/contacts.yaml for contact info,
// backend/config/about/bio.txt (bio.en.txt) for the about-page text and
// backend/config/about/faq.yaml for the questions and answers on the site.
package about

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"design-profile/backend/internal/model"

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

// LoadFAQ reads the FAQ from the YAML file at path. A missing or empty file
// yields an empty list. A malformed file, an unknown key (e.g. a typo) or an
// item without Russian text is an error: the server refuses to start rather
// than show a broken FAQ.
func LoadFAQ(path string) ([]model.FAQItem, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []model.FAQItem{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read faq file %q: %w", path, err)
	}

	items := []model.FAQItem{}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&items); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("parse faq file %q: %w", path, err)
	}

	for i := range items {
		item := &items[i]
		for _, text := range []*model.LocalizedText{&item.Question, &item.Answer} {
			text.RU = strings.TrimSpace(text.RU)
			text.EN = strings.TrimSpace(text.EN)
		}
		if item.Question.RU == "" || item.Answer.RU == "" {
			return nil, fmt.Errorf("faq file %q: item %d needs both question.ru and answer.ru", path, i+1)
		}
	}
	return items, nil
}
