package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"design-profile/backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getFAQ(t *testing.T, items []model.FAQItem) (int, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/faq", NewFAQHandler(items).ListFAQ)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/faq", nil))
	return w.Code, w.Body.String()
}

func TestListFAQ(t *testing.T) {
	t.Run("returns items in order", func(t *testing.T) {
		items := []model.FAQItem{
			{Question: model.LocalizedText{RU: "Вопрос 1", EN: "Question 1"}, Answer: model.LocalizedText{RU: "Ответ 1"}},
			{Question: model.LocalizedText{RU: "Вопрос 2"}, Answer: model.LocalizedText{RU: "Ответ 2"}},
		}
		code, body := getFAQ(t, items)
		assert.Equal(t, http.StatusOK, code)

		var resp struct {
			Data []model.FAQItem `json:"data"`
		}
		require.NoError(t, json.Unmarshal([]byte(body), &resp))
		assert.Equal(t, items, resp.Data)
	})

	t.Run("no FAQ is an empty list, not null", func(t *testing.T) {
		code, body := getFAQ(t, nil)
		assert.Equal(t, http.StatusOK, code)
		assert.JSONEq(t, `{"data":[]}`, body)
	})
}
