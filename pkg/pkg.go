package pkg

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type AvailableCompute struct {
	TotalCPUModel string  `json:"total_cpu_model"`
	TotalCPUCores int     `json:"total_cpu_cores"`
	ToalCPUGhz    float64 `json:"total_cpu_ghz"`
	TotalRAM      float64 `json:"total_ram"`

	FreeCPUCores int     `json:"free_cpu_cores"`
	FreeRAM      float64 `json:"free_ram"`
}

func GetComputeAvailable() (*AvailableCompute, error) {
	// Get CPU information
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting CPU information")
	}

	// Get memory information
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting memory information")
	}
	// Calculate total RAM in Gigabytes
	totalRAM := float64(vmem.Total) / 1024 / 1024 / 1024
	freeRAM := float64(vmem.Free) / 1024 / 1024 / 1024

	// Calculate total CPU speed in GHz
	totalCPUGhz := cpuInfo[0].Mhz / 1000

	return &AvailableCompute{
		TotalCPUModel: cpuInfo[0].ModelName,
		TotalCPUCores: int(cpuInfo[0].Cores),
		ToalCPUGhz:    totalCPUGhz,
		TotalRAM:      totalRAM,

		FreeCPUCores: runtime.NumCPU(),
		FreeRAM:      freeRAM,
	}, nil
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

// RunCmd executes the given command with the provided arguments
func RunCmd(name string, args ...string) ([]string, int, error) {

	fmt.Printf("Executing command: %s %s\n", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)

	// get the outputs
	var outputs []string

	// Attach the stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, 0, fmt.Errorf("error attaching stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, 0, fmt.Errorf("error attaching stderr pipe: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	// get the outputs
	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := stdout.Read(buf)
			if n > 0 {
				outputs = append(outputs, "Info: "+strings.TrimSpace(string(buf[:n])))
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := stderr.Read(buf)
			if n > 0 {
				outputs = append(outputs, "Error: "+strings.TrimSpace(string(buf[:n])))
			}
			if err != nil {
				break
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, 0, fmt.Errorf("error starting command: %w", err)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		return outputs, 0, fmt.Errorf("error waiting for command to finish: %w", err)
	}

	select {
	case <-time.After(time.Minute / 2):
		cmd.Process.Kill()
	default:
		wg.Wait()
	}

	fmt.Println("Command executed successfully")
	return outputs, cmd.ProcessState.Pid(), nil
}
