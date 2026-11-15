package handler

import (
	"design-profile/backend/internal/model"

	"github.com/gin-gonic/gin"
)

// FAQHandler serves the FAQ loaded from config/about/faq.yaml at startup.
type FAQHandler struct {
	items []model.FAQItem
}

func NewFAQHandler(items []model.FAQItem) *FAQHandler {
	if items == nil {
		items = []model.FAQItem{}
	}
	return &FAQHandler{items: items}
}

// ListFAQ godoc
// @Summary      List FAQ
// @Description  Returns the questions and answers shown on the site, in the order of config/about/faq.yaml. Each text has Russian ("ru") and, optionally, English ("en").
// @Tags         content
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /faq [get]
func (h *FAQHandler) ListFAQ(c *gin.Context) {
	ok(c, h.items)
}
