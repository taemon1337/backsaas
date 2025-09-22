package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Storage interface for persisting coverage data
type Storage interface {
	StoreCoverage(service string, data *CoverageData) error
	GetCoverage(service string) (*CoverageData, error)
	GetAllCoverage() (map[string]*CoverageData, error)
	GetCoverageHistory(service string, limit int) ([]*CoverageData, error)
	GetSummary() (*Summary, error)
}

// CoverageData represents comprehensive service health information
type CoverageData struct {
	Service      string                 `json:"service"`
	Timestamp    time.Time              `json:"timestamp"`
	Overall      float64                `json:"overall"`
	Packages     map[string]float64     `json:"packages"`
	Functions    map[string]float64     `json:"functions"`
	Lines        int                    `json:"lines"`
	CoveredLines int                    `json:"covered_lines"`
	Files        []FileData             `json:"files"`
	TestResults  *TestResults           `json:"test_results,omitempty"`
	ServiceHealth *ServiceHealth        `json:"service_health,omitempty"`
	TestStatus   *TestStatus            `json:"test_status,omitempty"`
	IntegrationTests *IntegrationTestStatus `json:"integration_tests,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// FileData represents coverage information for a single file
type FileData struct {
	Name         string  `json:"name"`
	Lines        int     `json:"lines"`
	CoveredLines int     `json:"covered_lines"`
	Coverage     float64 `json:"coverage"`
}

// TestResults represents test execution results
type TestResults struct {
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Skipped  int           `json:"skipped"`
	Duration time.Duration `json:"duration"`
	Output   string        `json:"output,omitempty"`
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Status      string    `json:"status"`      // "healthy", "unhealthy", "unknown"
	LastCheck   time.Time `json:"last_check"`
	ResponseTime int64    `json:"response_time_ms"`
	Endpoint    string    `json:"endpoint,omitempty"`
	Error       string    `json:"error,omitempty"`
}

// TestStatus represents the testing status of a service
type TestStatus struct {
	UnitTests        *TestSuite `json:"unit_tests"`
	IntegrationTests *TestSuite `json:"integration_tests"`
	E2ETests         *TestSuite `json:"e2e_tests,omitempty"`
	TotalComponents  int        `json:"total_components"`
	TestedComponents int        `json:"tested_components"`
	TestCoverage     float64    `json:"test_coverage_percent"`
}

// TestSuite represents a collection of tests
type TestSuite struct {
	Exists   bool      `json:"exists"`
	Count    int       `json:"count"`
	Passed   int       `json:"passed"`
	Failed   int       `json:"failed"`
	LastRun  time.Time `json:"last_run"`
	Status   string    `json:"status"` // "passing", "failing", "unknown", "missing"
}

// IntegrationTestStatus represents E2E integration test status
type IntegrationTestStatus struct {
	LastRun          time.Time              `json:"last_run"`
	Status           string                 `json:"status"` // "running", "passed", "failed", "not_run"
	TotalMilestones  int                    `json:"total_milestones"`
	PassedMilestones int                    `json:"passed_milestones"`
	FailedMilestones int                    `json:"failed_milestones"`
	Duration         time.Duration          `json:"duration"`
	Milestones       []MilestoneResult      `json:"milestones"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
}

// MilestoneResult represents individual milestone test result
type MilestoneResult struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Status      string        `json:"status"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	PassedCount int           `json:"passed_count"`
	FailedCount int           `json:"failed_count"`
}

// Summary represents overall coverage summary
type Summary struct {
	Timestamp      time.Time             `json:"timestamp"`
	OverallCoverage float64              `json:"overall_coverage"`
	Services       map[string]float64    `json:"services"`
	TotalLines     int                   `json:"total_lines"`
	CoveredLines   int                   `json:"covered_lines"`
	TestsPassed    int                   `json:"tests_passed"`
	TestsFailed    int                   `json:"tests_failed"`
	Trends         map[string][]float64  `json:"trends"`
}

// MemoryStorage implements in-memory storage
type MemoryStorage struct {
	mu       sync.RWMutex
	data     map[string][]*CoverageData
	maxItems int
}

// FileStorage implements file-based storage
type FileStorage struct {
	mu       sync.RWMutex
	basePath string
	maxItems int
}

// New creates a new storage instance
func New(storageType string, config map[string]string) (Storage, error) {
	switch storageType {
	case "memory":
		return &MemoryStorage{
			data:     make(map[string][]*CoverageData),
			maxItems: 100,
		}, nil
	case "file":
		path := config["path"]
		if path == "" {
			path = "/tmp/coverage-data"
		}
		return &FileStorage{
			basePath: path,
			maxItems: 100,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// MemoryStorage implementation
func (m *MemoryStorage) StoreCoverage(service string, data *CoverageData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.data[service] == nil {
		m.data[service] = make([]*CoverageData, 0)
	}

	// Add new data
	m.data[service] = append(m.data[service], data)

	// Keep only the latest maxItems
	if len(m.data[service]) > m.maxItems {
		m.data[service] = m.data[service][len(m.data[service])-m.maxItems:]
	}

	return nil
}

func (m *MemoryStorage) GetCoverage(service string) (*CoverageData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := m.data[service]
	if len(data) == 0 {
		return nil, fmt.Errorf("no coverage data found for service: %s", service)
	}

	return data[len(data)-1], nil
}

func (m *MemoryStorage) GetAllCoverage() (map[string]*CoverageData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*CoverageData)
	for service, data := range m.data {
		if len(data) > 0 {
			result[service] = data[len(data)-1]
		}
	}

	return result, nil
}

func (m *MemoryStorage) GetCoverageHistory(service string, limit int) ([]*CoverageData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := m.data[service]
	if len(data) == 0 {
		return []*CoverageData{}, nil
	}

	start := 0
	if len(data) > limit {
		start = len(data) - limit
	}

	return data[start:], nil
}

func (m *MemoryStorage) GetSummary() (*Summary, error) {
	allCoverage, err := m.GetAllCoverage()
	if err != nil {
		return nil, err
	}

	summary := &Summary{
		Timestamp: time.Now(),
		Services:  make(map[string]float64),
		Trends:    make(map[string][]float64),
	}

	var totalLines, coveredLines, testsPassed, testsFailed int
	var totalCoverage float64
	serviceCount := 0

	for service, data := range allCoverage {
		summary.Services[service] = data.Overall
		totalLines += data.Lines
		coveredLines += data.CoveredLines
		totalCoverage += data.Overall
		serviceCount++

		if data.TestResults != nil {
			testsPassed += data.TestResults.Passed
			testsFailed += data.TestResults.Failed
		}

		// Get trend data (last 10 points)
		history, _ := m.GetCoverageHistory(service, 10)
		trends := make([]float64, len(history))
		for i, h := range history {
			trends[i] = h.Overall
		}
		summary.Trends[service] = trends
	}

	if serviceCount > 0 {
		summary.OverallCoverage = totalCoverage / float64(serviceCount)
	}
	summary.TotalLines = totalLines
	summary.CoveredLines = coveredLines
	summary.TestsPassed = testsPassed
	summary.TestsFailed = testsFailed

	return summary, nil
}

// FileStorage implementation
func (f *FileStorage) StoreCoverage(service string, data *CoverageData) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Ensure directory exists
	serviceDir := filepath.Join(f.basePath, service)
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return err
	}

	// Store as JSON file with timestamp
	filename := fmt.Sprintf("%d.json", data.Timestamp.Unix())
	filepath := filepath.Join(serviceDir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *FileStorage) GetCoverage(service string) (*CoverageData, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	serviceDir := filepath.Join(f.basePath, service)
	files, err := os.ReadDir(serviceDir)
	if err != nil {
		return nil, fmt.Errorf("no coverage data found for service: %s", service)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no coverage data found for service: %s", service)
	}

	// Sort files by name (timestamp) and get the latest
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})

	latestFile := filepath.Join(serviceDir, files[0].Name())
	file, err := os.Open(latestFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data CoverageData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (f *FileStorage) GetAllCoverage() (map[string]*CoverageData, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make(map[string]*CoverageData)

	entries, err := os.ReadDir(f.basePath)
	if err != nil {
		return result, nil // Return empty result if directory doesn't exist
	}

	for _, entry := range entries {
		if entry.IsDir() {
			service := entry.Name()
			if data, err := f.GetCoverage(service); err == nil {
				result[service] = data
			}
		}
	}

	return result, nil
}

func (f *FileStorage) GetCoverageHistory(service string, limit int) ([]*CoverageData, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	serviceDir := filepath.Join(f.basePath, service)
	files, err := os.ReadDir(serviceDir)
	if err != nil {
		return []*CoverageData{}, nil
	}

	// Sort files by name (timestamp) in descending order
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})

	// Limit the number of files to read
	if len(files) > limit {
		files = files[:limit]
	}

	var history []*CoverageData
	for _, file := range files {
		filePath := filepath.Join(serviceDir, file.Name())
		f, err := os.Open(filePath)
		if err != nil {
			continue
		}

		var data CoverageData
		if err := json.NewDecoder(f).Decode(&data); err != nil {
			f.Close()
			continue
		}
		f.Close()

		history = append(history, &data)
	}

	// Reverse to get chronological order
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

func (f *FileStorage) GetSummary() (*Summary, error) {
	allCoverage, err := f.GetAllCoverage()
	if err != nil {
		return nil, err
	}

	summary := &Summary{
		Timestamp: time.Now(),
		Services:  make(map[string]float64),
		Trends:    make(map[string][]float64),
	}

	var totalLines, coveredLines, testsPassed, testsFailed int
	var totalCoverage float64
	serviceCount := 0

	for service, data := range allCoverage {
		summary.Services[service] = data.Overall
		totalLines += data.Lines
		coveredLines += data.CoveredLines
		totalCoverage += data.Overall
		serviceCount++

		if data.TestResults != nil {
			testsPassed += data.TestResults.Passed
			testsFailed += data.TestResults.Failed
		}

		// Get trend data (last 10 points)
		history, _ := f.GetCoverageHistory(service, 10)
		trends := make([]float64, len(history))
		for i, h := range history {
			trends[i] = h.Overall
		}
		summary.Trends[service] = trends
	}

	if serviceCount > 0 {
		summary.OverallCoverage = totalCoverage / float64(serviceCount)
	}
	summary.TotalLines = totalLines
	summary.CoveredLines = coveredLines
	summary.TestsPassed = testsPassed
	summary.TestsFailed = testsFailed

	return summary, nil
}
