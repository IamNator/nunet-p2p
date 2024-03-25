package pkg

import (
	"runtime"

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
