package audit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Middleware struct {
	svc *Service
}

func NewMiddleware(svc *Service) *Middleware {
	return &Middleware{svc: svc}
}

func (m *Middleware) Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		userID, _ := c.Get("userID")
		uid, _ := userID.(uint)

		action := c.Request.Method + " " + c.FullPath()
		module := "api"
		description := c.Request.URL.Path
		ip := c.ClientIP()
		ua := c.Request.UserAgent()

		if uid > 0 {
			go m.svc.LogWithIP(uid, action, module, description, ip, ua)
		}

		_ = start
	}
}

func RegisterRoutes(r *gin.RouterGroup, svc *Service) {
	audit := r.Group("/audit")
	{
		audit.GET("/logs", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
			action := c.Query("action")
			module := c.Query("module")

			query := &AuditQuery{
				UserID:   uid,
				Action:    action,
				Module:    module,
				Page:      page,
				PageSize:  pageSize,
			}

			if start := c.Query("start_time"); start != "" {
				if t, err := time.Parse(time.RFC3339, start); err == nil {
					query.StartTime = t
				}
			}

			if end := c.Query("end_time"); end != "" {
				if t, err := time.Parse(time.RFC3339, end); err == nil {
					query.EndTime = t
				}
			}

			result, err := svc.Query(query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, result)
		})
	}
}
