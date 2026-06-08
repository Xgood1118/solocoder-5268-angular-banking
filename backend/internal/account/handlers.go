package account

import (
	"banking/pkg/masking"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, svc *Service) {
	accounts := r.Group("/accounts")
	{
		accounts.POST("", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			var req OpenAccountRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			req.UserID = uid

			account, err := svc.OpenAccount(&req)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			account.AccountNumber = masking.MaskCardNumber(account.AccountNumber)

			c.JSON(http.StatusCreated, account)
		})

		accounts.GET("", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			accounts, err := svc.ListAccounts(uid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			for i := range accounts {
				accounts[i].AccountNumber = masking.MaskCardNumber(accounts[i].AccountNumber)
			}

			c.JSON(http.StatusOK, accounts)
		})

		accounts.GET("/:id", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			id, _ := strconv.Atoi(c.Param("id"))

			account, err := svc.GetAccount(uid, uint(id))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			account.AccountNumber = masking.MaskCardNumber(account.AccountNumber)

			c.JSON(http.StatusOK, account)
		})

		accounts.POST("/:id/freeze", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			id, _ := strconv.Atoi(c.Param("id"))

			var body struct {
				Reason string `json:"reason"`
			}
			c.ShouldBindJSON(&body)

			if err := svc.FreezeAccount(uid, uint(id), body.Reason); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "account frozen"})
		})

		accounts.POST("/:id/unfreeze", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			id, _ := strconv.Atoi(c.Param("id"))

			if err := svc.UnfreezeAccount(uid, uint(id)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "account unfrozen"})
		})

		accounts.DELETE("/:id", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			id, _ := strconv.Atoi(c.Param("id"))

			if err := svc.CloseAccount(uid, uint(id)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "account closed"})
		})

		accounts.GET("/:id/ledger", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			id, _ := strconv.Atoi(c.Param("id"))
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

			account, err := svc.GetAccount(uid, uint(id))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			_ = account

			total, entries, err := svc.GetLedger(uint(id), page, pageSize)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"total":     total,
				"page":      page,
				"page_size": pageSize,
				"entries":   entries,
			})
		})
	}
}
