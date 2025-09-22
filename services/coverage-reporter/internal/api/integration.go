package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/backsaas/platform/coverage-reporter/internal/storage"
)

// IntegrationTestRequest represents incoming integration test data
type IntegrationTestRequest struct {
	MilestoneID string        `json:"milestone_id"`
	Name        string        `json:"name"`
	Status      string        `json:"status"`
	Duration    int64         `json:"duration"` // milliseconds
	Passed      int           `json:"passed"`
	Failed      int           `json:"failed"`
	Timestamp   int64         `json:"timestamp"`
}

// HandleIntegrationTestReport handles integration test milestone reports
func (s *Server) HandleIntegrationTestReport(c *gin.Context) {
	var req IntegrationTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current system-wide integration test status
	systemData, err := s.storage.GetCoverage("system")
	if err != nil {
		// Create new system entry if it doesn't exist
		systemData = &storage.CoverageData{
			Service:   "system",
			Timestamp: time.Now(),
			IntegrationTests: &storage.IntegrationTestStatus{
				Status:      "not_run",
				Milestones:  []storage.MilestoneResult{},
			},
		}
	}

	// Initialize integration tests if nil
	if systemData.IntegrationTests == nil {
		systemData.IntegrationTests = &storage.IntegrationTestStatus{
			Status:     "not_run",
			Milestones: []storage.MilestoneResult{},
		}
	}

	// Update milestone result
	milestone := storage.MilestoneResult{
		ID:          req.MilestoneID,
		Name:        req.Name,
		Status:      req.Status,
		StartTime:   time.Unix(req.Timestamp, 0),
		EndTime:     time.Unix(req.Timestamp, 0).Add(time.Duration(req.Duration) * time.Millisecond),
		Duration:    time.Duration(req.Duration) * time.Millisecond,
		PassedCount: req.Passed,
		FailedCount: req.Failed,
	}

	// Update or add milestone
	found := false
	for i, existing := range systemData.IntegrationTests.Milestones {
		if existing.ID == req.MilestoneID {
			systemData.IntegrationTests.Milestones[i] = milestone
			found = true
			break
		}
	}
	if !found {
		systemData.IntegrationTests.Milestones = append(systemData.IntegrationTests.Milestones, milestone)
	}

	// Update overall integration test status
	totalMilestones := len(systemData.IntegrationTests.Milestones)
	passedMilestones := 0
	failedMilestones := 0
	var totalDuration time.Duration

	for _, m := range systemData.IntegrationTests.Milestones {
		if m.Status == "passed" {
			passedMilestones++
		} else if m.Status == "failed" {
			failedMilestones++
		}
		totalDuration += m.Duration
	}

	systemData.IntegrationTests.TotalMilestones = totalMilestones
	systemData.IntegrationTests.PassedMilestones = passedMilestones
	systemData.IntegrationTests.FailedMilestones = failedMilestones
	systemData.IntegrationTests.Duration = totalDuration
	systemData.IntegrationTests.LastRun = time.Now()

	// Determine overall status
	if failedMilestones > 0 {
		systemData.IntegrationTests.Status = "failed"
	} else if passedMilestones == totalMilestones && totalMilestones > 0 {
		systemData.IntegrationTests.Status = "passed"
	} else {
		systemData.IntegrationTests.Status = "running"
	}

	// Store updated data
	if err := s.storage.StoreCoverage("system", systemData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store integration test data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Integration test milestone recorded",
		"milestone": req.MilestoneID,
		"status": req.Status,
		"overall_status": systemData.IntegrationTests.Status,
	})
}

// HandleGetIntegrationTests returns current integration test status
func (s *Server) HandleGetIntegrationTests(c *gin.Context) {
	systemData, err := s.storage.GetCoverage("system")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No integration test data found"})
		return
	}

	if systemData.IntegrationTests == nil {
		c.JSON(http.StatusOK, gin.H{
			"status": "not_run",
			"milestones": []storage.MilestoneResult{},
		})
		return
	}

	c.JSON(http.StatusOK, systemData.IntegrationTests)
}
