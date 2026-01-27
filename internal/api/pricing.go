package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/railzwaylabs/railzway-cloud/pkg/railzwayclient"
	"go.uber.org/zap"
)

// ListPrices proxies the request to Railzway OSS
func (r *Router) ListPrices(c *gin.Context) {
	productID := c.Query("product_id")
	code := c.Query("code")
	var opts *railzwayclient.PriceListOptions
	if code != "" {
		opts = &railzwayclient.PriceListOptions{Code: code}
	}
	prices, err := r.client.ListPrices(c.Request.Context(), opts)
	if err != nil {
		r.logger.Error("failed to list prices", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch prices"})
		return
	}
	if productID != "" {
		filtered := make([]railzwayclient.Price, 0, len(prices))
		for _, price := range prices {
			if price.ProductID == productID {
				filtered = append(filtered, price)
			}
		}
		prices = filtered
	}
	c.JSON(http.StatusOK, gin.H{"data": prices})
}

// ListPriceAmounts proxies the request to Railzway OSS
func (r *Router) ListPriceAmounts(c *gin.Context) {
	priceID := c.Query("price_id")
	amounts, err := r.client.ListPriceAmounts(c.Request.Context(), priceID)
	if err != nil {
		r.logger.Error("failed to list price amounts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch price amounts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": amounts})
}
