package httputils

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/mailru/easyjson/jwriter"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type SimplifiedRuntimeMemStats struct {
	Alloc        float64 `json:"alloc"`
	TotalAlloc   float64 `json:"totalAlloc"`
	Sys          float64 `json:"sys"`
	Lookups      float64 `json:"lookups"`
	Mallocs      float64 `json:"mallocs"`
	Frees        float64 `json:"frees"`
	HeapAlloc    float64 `json:"heapAlloc"`
	HeapSys      float64 `json:"heapSys"`
	HeapIdle     float64 `json:"heapIdle"`
	HeapInuse    float64 `json:"heapInuse"`
	HeapReleased float64 `json:"heapReleased"`
	HeapObjects  float64 `json:"heapObjects"`
	StackInuse   float64 `json:"stackInuse"`
	StackSys     float64 `json:"stackSys"`
	MSpanInuse   float64 `json:"mSpanInuse"`
	MSpanSys     float64 `json:"mSpanSys"`
}

type Memory struct {
	Id                 string                    `json:"id"`
	MemTotal           int                       `json:"total"`
	MemFree            int                       `json:"free"`
	MemAvailable       int                       `json:"available"`
	RuntimeMemoryStats SimplifiedRuntimeMemStats `json:"runtimeMemoryStats"`
}

type PsEntry struct {
	Pid    int    `json:"pid"`
	User   string `json:"user"`
	VmRss  int64  `json:"vmRss"`
	VmSize int64  `json:"vmSize"`
	Name   string `json:"name"`
}

func (m Memory) MarshalEasyJSON(w *jwriter.Writer) {
	bytes, err := json.Marshal(m)
	w.Raw(bytes, err)
}

func getAllMemoryStats() (Memory, error) {
	stats, err := ReadMemoryStats()
	if err != nil {
		fmt.Printf("Error reading memory stats from /proc/meminfo: %v", err)
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats.RuntimeMemoryStats = SimplifiedRuntimeMemStats{
		Alloc:        float64(m.Alloc) / 1024.0,
		TotalAlloc:   float64(m.TotalAlloc) / 1024.0,
		Sys:          float64(m.Sys) / 1024.0,
		Lookups:      float64(m.Lookups) / 1024.0,
		Mallocs:      float64(m.Mallocs) / 1024.0,
		Frees:        float64(m.Frees) / 1024.0,
		HeapAlloc:    float64(m.HeapAlloc) / 1024.0,
		HeapSys:      float64(m.HeapSys) / 1024.0,
		HeapIdle:     float64(m.HeapIdle) / 1024.0,
		HeapInuse:    float64(m.HeapInuse) / 1024.0,
		HeapReleased: float64(m.HeapReleased) / 1024.0,
		HeapObjects:  float64(m.HeapObjects) / 1024.0,
		StackInuse:   float64(m.StackInuse) / 1024.0,
		StackSys:     float64(m.StackSys) / 1024.0,
		MSpanInuse:   float64(m.MSpanInuse) / 1024.0,
		MSpanSys:     float64(m.MSpanSys) / 1024.0,
	}
	if ProcessHash == "" {
		ProcessHash = getRandomProcessHash4bytes()
	}
	stats.Id = ProcessHash
	return stats, err
}

func ReadMemoryStats() (memoryStats Memory, err error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v", err)
		}
	}(file)
	bufio.NewScanner(file)
	scanner := bufio.NewScanner(file)
	memoryStats = Memory{}
	for scanner.Scan() {
		key, value := parseLine(scanner.Text())
		switch key {
		case "MemTotal":
			memoryStats.MemTotal = value
		case "MemFree":
			memoryStats.MemFree = value
		case "MemAvailable":
			memoryStats.MemAvailable = value
		}
	}
	return
}

func parseLine(raw string) (key string, value int) {
	//fmt.Println(raw)
	text := strings.ReplaceAll(raw[:len(raw)-2], " ", "")
	keyValue := strings.Split(text, ":")
	return keyValue[0], toInt(keyValue[1])
}

func toInt(raw string) int {
	if raw == "" {
		return 0
	}
	res, err := strconv.Atoi(raw)
	if err != nil {
		panic(err)
	}
	return res
}

var RegexPid = regexp.MustCompile(`^\d+$`)

func parseProcessList() (out []PsEntry, err error) {
	// List /proc/\d+
	processes, err := os.ReadDir("/proc")
	if err != nil {
		return out, err
	}
	for _, process := range processes {
		possiblePid := process.Name()
		if RegexPid.MatchString(possiblePid) {
			// Read /proc/\d+/status
			status, err := os.ReadFile("/proc/" + possiblePid + "/status")
			if err != nil {
				return out, err
			}
			// Parse /proc/\d+/status
			var entry PsEntry
			entry.Pid, err = strconv.Atoi(possiblePid)
			// read lines
			lines := bytes.Split(status, []byte{'\n'})
			for _, line := range lines {
				// read fields
				fields := bytes.Split(line, []byte{':'})
				switch string(bytes.TrimSpace(fields[0])) {
				case "VmRSS":
					splt := bytes.Split(bytes.TrimSpace(fields[1]), []byte{' '})
					n, err := strconv.Atoi(string(bytes.TrimSpace(splt[0])))
					if err != nil {
						return out, fmt.Errorf("error parsing VmRss: %w", err)
					}
					entry.VmRss = int64(n)
				case "VmSize":
					splt := bytes.Split(bytes.TrimSpace(fields[1]), []byte{' '})
					n, err := strconv.Atoi(string(bytes.TrimSpace(splt[0])))
					if err != nil {
						return out, fmt.Errorf("error parsing VmSize: %w", err)
					}
					entry.VmSize = int64(n)
				case "Name":
					entry.Name = string(bytes.TrimSpace(fields[1]))
				}
			}
			out = append(out, entry)
		}
	}
	return
}
