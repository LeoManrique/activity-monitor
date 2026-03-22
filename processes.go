package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type ProcessInfo struct {
	PID     int
	RSS     int64
	Command string
}

func parseProcessLine(line string) (ProcessInfo, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return ProcessInfo{}, fmt.Errorf("got a ps process with less than 3 fields:\n" + line)
	}
	pid, err1 := strconv.Atoi(fields[0])
	if err1 != nil {
		return ProcessInfo{}, fmt.Errorf("invalid pid: " + fields[0])
	}
	rss, err2 := strconv.Atoi(fields[1])
	if err2 != nil {
		return ProcessInfo{}, fmt.Errorf("invalid rss: " + fields[1])
	}
	psCommand := strings.Join(fields[2:], " ")
	psProgram := filepath.Base(psCommand)
	return ProcessInfo{pid, int64(rss), psProgram}, nil
}

func parseAllProcesses(output string) []ProcessInfo {
	var result []ProcessInfo
	splitOutput := strings.Split(output, "\n")
	for index, value := range splitOutput {
		if index == 0 {
			continue
		}
		processInfo, err := parseProcessLine(value)
		if err != nil {
			// skipped value
		}
		result = append(result, processInfo)
	}
	return result
}

func getProcessList() (string, error) {
	output, err := exec.Command("ps",
		"-axo",
		"pid,rss,comm",
	).Output()
	if err != nil {
		return "", err
	}
	return string(output), err
}
