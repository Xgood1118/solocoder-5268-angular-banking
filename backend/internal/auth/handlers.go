package auth

import (
	"banking/pkg/masking"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, svc *Service) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", func(c *gin.Context) {
			var req RegisterRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			user, err := svc.Register(&req)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			user.IDCard = masking.MaskIDCard(user.IDCard)
			user.Phone = masking.MaskPhone(user.Phone)
			user.Email = masking.MaskEmail(user.Email)

			c.JSON(http.StatusCreated, user)
		})

		auth.POST("/login", func(c *gin.Context) {
			var req LoginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			resp, err := svc.Login(&req)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}

			if resp.User != nil {
				resp.User.IDCard = masking.MaskIDCard(resp.User.IDCard)
				resp.User.Phone = masking.MaskPhone(resp.User.Phone)
				resp.User.Email = masking.MaskEmail(resp.User.Email)
			}

			c.JSON(http.StatusOK, resp)
		})

		auth.POST("/twofa/verify", func(c *gin.Context) {
			token := c.Query("token")
			var req TwoFARequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			resp, err := svc.VerifyTwoFA(token, req.Code)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}

			if resp.User != nil {
				resp.User.IDCard = masking.MaskIDCard(resp.User.IDCard)
				resp.User.Phone = masking.MaskPhone(resp.User.Phone)
				resp.User.Email = masking.MaskEmail(resp.User.Email)
			}

			c.JSON(http.StatusOK, resp)
		})
	}
}

func RegisterProtectedRoutes(r *gin.RouterGroup, svc *Service) {
	auth := r.Group("/auth")
	{
		auth.POST("/password", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			var req struct {
				OldPassword string `json:"old_password" binding:"required"`
				NewPassword string `json:"new_password" binding:"required,min=8"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := svc.ChangePassword(uid, req.OldPassword, req.NewPassword); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
		})

		auth.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			user, err := svc.GetUser(uid)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}

			user.IDCard = masking.MaskIDCard(user.IDCard)
			user.Phone = masking.MaskPhone(user.Phone)
			user.Email = masking.MaskEmail(user.Email)

			c.JSON(http.StatusOK, user)
		})
	}
}

func AuthMiddleware(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		userID, err := svc.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
