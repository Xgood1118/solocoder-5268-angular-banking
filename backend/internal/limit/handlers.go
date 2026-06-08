package limit

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, svc *Service) {
	limits := r.Group("/limits")
	{
		limits.GET("", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			scope := c.DefaultQuery("scope", "transfer")
			result := svc.GetLimits(uid, scope)

			c.JSON(http.StatusOK, result)
		})

		limits.POST("", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			var req struct {
				LimitType string  `json:"limit_type" binding:"required"`
				Amount    float64 `json:"amount" binding:"required,gt=0"`
				Scope     string  `json:"scope"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if req.Scope == "" {
				req.Scope = "transfer"
			}

			if err := svc.SetLimit(uid, LimitType(req.LimitType), req.Amount, req.Scope); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "limit updated"})
		})

		limits.GET("/configs", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			configs, err := svc.GetUserLimitConfigs(uid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, configs)
		})
	}
}
