package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
)

// ReportHandler exposes admin business-metrics reports.
type ReportHandler struct {
	reportService *service.ReportService
}

// NewReportHandler builds a ReportHandler backed by reportService.
func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// Dashboard handles GET /api/admin/reports/dashboard.
func (h *ReportHandler) Dashboard(c *gin.Context) {
	var query model.ReportPeriodQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		_ = c.Error(apperror.Validation("Parameter laporan tidak valid", bindingErrors(err)))
		return
	}

	report, err := h.reportService.Dashboard(c.Request.Context(), query.Days)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": report})
}
