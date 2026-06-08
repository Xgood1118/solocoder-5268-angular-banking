package transfer

import (
	"banking/pkg/masking"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, svc *Service) {
	transfers := r.Group("/transfers")
	{
		transfers.POST("", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			var req CreateTransferRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			req.UserID = uid

			transfer, err := svc.CreateTransfer(&req)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			transfer.FromAccountNo = masking.MaskCardNumber(transfer.FromAccountNo)
			transfer.ToAccountNo = masking.MaskCardNumber(transfer.ToAccountNo)

			c.JSON(http.StatusCreated, transfer)
		})

		transfers.GET("", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
			status := c.Query("status")

			total, transfers, err := svc.ListTransfers(uid, page, pageSize, status)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			for i := range transfers {
				transfers[i].FromAccountNo = masking.MaskCardNumber(transfers[i].FromAccountNo)
				transfers[i].ToAccountNo = masking.MaskCardNumber(transfers[i].ToAccountNo)
			}

			c.JSON(http.StatusOK, gin.H{
				"total":     total,
				"page":      page,
				"page_size": pageSize,
				"transfers": transfers,
			})
		})

		transfers.GET("/biz/:biz_id", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			bizID := c.Param("biz_id")

			transfer, err := svc.GetByBizID(uid, bizID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			transfer.FromAccountNo = masking.MaskCardNumber(transfer.FromAccountNo)
			transfer.ToAccountNo = masking.MaskCardNumber(transfer.ToAccountNo)

			c.JSON(http.StatusOK, transfer)
		})

		transfers.GET("/:id", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			uid, _ := userID.(uint)

			id, _ := strconv.Atoi(c.Param("id"))

			transfer, err := svc.GetTransfer(uid, uint(id))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			transfer.FromAccountNo = masking.MaskCardNumber(transfer.FromAccountNo)
			transfer.ToAccountNo = masking.MaskCardNumber(transfer.ToAccountNo)

			c.JSON(http.StatusOK, transfer)
		})
	}
}
