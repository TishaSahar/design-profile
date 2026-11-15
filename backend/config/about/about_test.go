package about

import (
	"os"
	"path/filepath"
	"testing"

	"design-profile/backend/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFAQ(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "faq.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// The FAQ committed to the repository must always load: this test fails CI
// before a broken faq.yaml can reach the deploy.
func TestLoadFAQ_RepositoryFile(t *testing.T) {
	items, err := LoadFAQ("faq.yaml")
	require.NoError(t, err)
	assert.NotEmpty(t, items)
	for i, item := range items {
		assert.NotEmpty(t, item.Question.EN, "item %d has no English question", i+1)
		assert.NotEmpty(t, item.Answer.EN, "item %d has no English answer", i+1)
	}
}

func TestLoadFAQ(t *testing.T) {
	t.Run("keeps order, trims text, English is optional", func(t *testing.T) {
		path := writeFAQ(t, `
- question:
    ru: "  Первый вопрос?  "
    en: First question?
  answer:
    ru: |-
      Первый абзац.

      Второй абзац.
- question:
    ru: Второй вопрос?
  answer:
    ru: Ответ только на русском.
`)
		items, err := LoadFAQ(path)
		require.NoError(t, err)
		assert.Equal(t, []model.FAQItem{
			{
				Question: model.LocalizedText{RU: "Первый вопрос?", EN: "First question?"},
				Answer:   model.LocalizedText{RU: "Первый абзац.\n\nВторой абзац."},
			},
			{
				Question: model.LocalizedText{RU: "Второй вопрос?"},
				Answer:   model.LocalizedText{RU: "Ответ только на русском."},
			},
		}, items)
	})

	t.Run("missing file is an empty FAQ", func(t *testing.T) {
		items, err := LoadFAQ(filepath.Join(t.TempDir(), "absent.yaml"))
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("file with comments only is an empty FAQ", func(t *testing.T) {
		items, err := LoadFAQ(writeFAQ(t, "# nothing yet\n"))
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	for name, content := range map[string]string{
		"malformed YAML":         "- question:\n    ru: Вопрос: с двоеточием без кавычек\n  answer: [",
		"typo in a key":          "- questoin:\n    ru: Вопрос?\n  answer:\n    ru: Ответ.\n",
		"missing Russian answer": "- question:\n    ru: Вопрос?\n  answer:\n    en: Answer only in English.\n",
		"empty Russian question": "- question:\n    ru: \"   \"\n  answer:\n    ru: Ответ.\n",
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			_, err := LoadFAQ(writeFAQ(t, content))
			assert.Error(t, err)
		})
	}
}
