package pkg

import (
	"fmt"
	"os"
	"strconv"

	"github.com/libp2p/go-libp2p/core/host"
)

func GetEnvOrDefault(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func PrintHostInfo(host host.Host) {
	fmt.Println("Host ID:", host.ID())
	for _, addr := range host.Addrs() {
		fmt.Printf("Listening on %s/p2p/%s\n", addr, host.ID())
	}
}
