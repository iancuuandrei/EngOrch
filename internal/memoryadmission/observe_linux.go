//go:build linux

package memoryadmission

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func observe() Observation {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return unavailable()
	}
	defer file.Close()
	var availableKiB, totalKiB int64
	var foundAvailable, foundTotal bool
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 64*1024)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		value, parseErr := strconv.ParseInt(fields[1], 10, 64)
		if parseErr != nil || value < 0 {
			continue
		}
		switch fields[0] {
		case "MemAvailable:":
			availableKiB = value
			foundAvailable = true
		case "MemTotal:":
			totalKiB = value
			foundTotal = true
		}
	}
	if scanner.Err() != nil || !foundAvailable || !foundTotal || totalKiB == 0 {
		return unavailable()
	}
	return boundedObservation(SourceLinuxProcMeminfo, availableKiB/1024, totalKiB/1024)
}
