package recon

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, svc *Service) {
	recon := r.Group("/recon")
	{
		recon.GET("/reports", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			_, _ = userID.(uint)

			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
			accountID, _ := strconv.Atoi(c.DefaultQuery("account_id", "0"))

			total, reports, err := svc.GetReports(uint(accountID), page, pageSize)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"total":     total,
				"page":      page,
				"page_size": pageSize,
				"reports":   reports,
			})
		})

		recon.GET("/reports/:id/differences", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

			total, diffs, err := svc.GetDifferences(uint(id), page, pageSize)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"total":       total,
				"page":        page,
				"page_size":   pageSize,
				"differences": diffs,
			})
		})

		recon.POST("/trigger", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			_, _ = userID.(uint)

			var req struct {
				AccountID uint `json:"account_id" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			report, err := svc.TriggerRecon(req.AccountID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, report)
		})

		recon.POST("/differences/:id/resolve", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			id, _ := strconv.Atoi(c.Param("id"))

			if err := svc.ManualReconcile(uint(id), uid); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "reconciled"})
		})
	}
}
