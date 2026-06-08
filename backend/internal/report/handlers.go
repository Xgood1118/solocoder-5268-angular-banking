package report

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, svc *ReportService) {
	report := r.Group("/report")
	{
		report.GET("/balance/daily", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))

			result, err := svc.GetDailyBalanceReport(uid, date)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, result)
		})

		report.GET("/balance/weekly", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			result, err := svc.GetWeeklyBalanceReport(uid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, result)
		})

		report.GET("/transactions", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			startDate := c.Query("start_date")
			endDate := c.Query("end_date")

			if startDate == "" || endDate == "" {
				endDate = time.Now().Format("2006-01-02")
				startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
			}

			result, err := svc.GetTransactionReport(uid, startDate, endDate)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, result)
		})
	}
}
