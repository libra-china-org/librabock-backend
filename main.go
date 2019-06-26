package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func main() {
	db := NewDataBaseAdapter("root:test@(127.0.0.1:32773)/libra")
	db.migration()

	r := gin.Default()

	r.GET("/version", func(c *gin.Context) {
		offsetStr := c.DefaultQuery("offset", "0")
		limitStr := c.DefaultQuery("limit", "20")

		offset, err1 := strconv.Atoi(offsetStr)
		limit, err2 := strconv.Atoi(limitStr)

		if err1 != nil || err2 != nil {
			c.JSON(400, gin.H{"message": "bad request"})
			return
		}

		c.JSON(200, db.GetVersions(offset, limit))
	})

	r.GET("/version/:id", func(c *gin.Context) {
		id := c.Param("id")
		id64, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			c.JSON(400, gin.H{"message": "bad request"})
			return
		}

		version := db.GetVersion(uint64(id64))
		if version.ID == 0 {
			c.JSON(404, gin.H{"message": "not found"})
		} else {
			c.JSON(200, version)
		}
	})

	r.GET("/account/:address", func(c *gin.Context) {
		address := c.Param("address")
		_, err := hexToBytes(address)

		if len(address) != 64 || err != nil {
			c.JSON(400, gin.H{"message": "bad request"})
			return
		}

		rpc := NewLibraRPC(nil)
		r, err := rpc.GetAccountState(address)

		if err != nil {
			c.JSON(400, gin.H{"message": "bad request"})
		} else {
			if r == nil {
				c.JSON(404, gin.H{"message": "not found"})
			} else {
				c.JSON(200, r)
			}
		}
	})

	r.Run()
}