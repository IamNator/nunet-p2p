package pkg

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func GetComputeAvailable() (cpuAvailable int, ramAvailable float64, err error) {
	// Get CPU information
	cpuInfo, err := cpu.Info()
	if err != nil {
		return 0, 0, errors.Wrap(err, "Error getting CPU information")
	}

	// Get memory information
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, errors.Wrap(err, "Error getting memory information")
	}

	// Calculate total RAM in Gigabytes
	totalRAM := float64(vmem.Total) / 1024 / 1024 / 1024
	return int(cpuInfo[0].Cores), totalRAM, nil
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
