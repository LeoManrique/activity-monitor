package main

type AppMemoryGroup struct {
	Name         string `json:"name"`
	MemoryKB     int64  `json:"memoryKB"`
	ProcessCount int    `json:"processCount"`
}

type MemoryService struct{}

func (memoryservice *MemoryService) GetMemoryUsage() []AppMemoryGroup {
	chrome := AppMemoryGroup{
		Name:         "Google Chrome",
		MemoryKB:     1_500_000,
		ProcessCount: 23,
	}
	return []AppMemoryGroup{
		chrome,
		{Name: "Slack", MemoryKB: 800_000, ProcessCount: 5},
		{Name: "Firefox", MemoryKB: 600_000, ProcessCount: 12},
		{Name: "Finder", MemoryKB: 150_000, ProcessCount: 2},
	}
}
