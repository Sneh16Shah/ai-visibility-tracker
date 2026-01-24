package services

import (
	"context"
	"log"
	"time"

	"github.com/Sneh16Shah/ai-visibility-tracker/db"
)

// Scheduler handles scheduled analysis runs
type Scheduler struct {
	stopChan chan bool
	running  bool
}

// Global scheduler instance
var scheduler *Scheduler

// InitScheduler initializes and starts the scheduler
func InitScheduler() *Scheduler {
	scheduler = &Scheduler{
		stopChan: make(chan bool),
		running:  false,
	}
	return scheduler
}

// GetScheduler returns the global scheduler
func GetScheduler() *Scheduler {
	return scheduler
}

// Start begins the scheduler background goroutine
func (s *Scheduler) Start() {
	if s.running {
		return
	}
	s.running = true
	go s.run()
	log.Println("⏰ Scheduler started - checking for scheduled analyses every hour")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	if !s.running {
		return
	}
	s.stopChan <- true
	s.running = false
	log.Println("⏰ Scheduler stopped")
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkScheduledRuns()
		}
	}
}

// checkScheduledRuns checks all brands for scheduled analysis
func (s *Scheduler) checkScheduledRuns() {
	log.Println("⏰ Checking for scheduled analyses...")

	brandRepo := db.NewBrandRepository()
	brands, err := brandRepo.GetAllBrands()
	if err != nil {
		log.Printf("Error fetching brands for scheduling: %v", err)
		return
	}

	now := time.Now()

	for _, brand := range brands {
		if brand.ScheduleFrequency == "" || brand.ScheduleFrequency == "disabled" {
			continue
		}

		// Check if it's time to run
		shouldRun := false

		switch brand.ScheduleFrequency {
		case "daily":
			// Run if last run was more than 24 hours ago
			if brand.LastScheduledRun.IsZero() || now.Sub(brand.LastScheduledRun) > 24*time.Hour {
				shouldRun = true
			}
		case "weekly":
			// Run if last run was more than 7 days ago
			if brand.LastScheduledRun.IsZero() || now.Sub(brand.LastScheduledRun) > 7*24*time.Hour {
				shouldRun = true
			}
		}

		if shouldRun {
			log.Printf("⏰ Running scheduled analysis for brand: %s", brand.Name)
			s.runScheduledAnalysis(brand.ID)

			// Update last run time
			brandRepo.UpdateLastScheduledRun(brand.ID, now)
		}
	}

	// Also check for email alerts
	emailSvc := GetEmailService()
	if emailSvc != nil && emailSvc.IsEnabled() {
		emailSvc.CheckAndSendAlerts()
	}
}

// runScheduledAnalysis runs analysis for a brand
func (s *Scheduler) runScheduledAnalysis(brandID int) {
	analysisSvc := GetAnalysisService()
	if analysisSvc == nil {
		log.Println("Analysis service not initialized")
		return
	}

	// Get default prompts
	promptRepo := db.NewPromptRepository()
	prompts, err := promptRepo.GetAll()
	if err != nil || len(prompts) == 0 {
		log.Printf("No prompts available for scheduled analysis")
		return
	}

	// Use first 3 prompts
	promptIDs := make([]int, 0, 3)
	for i, p := range prompts {
		if i >= 3 {
			break
		}
		promptIDs = append(promptIDs, p.ID)
	}

	// Run analysis
	_, err = analysisSvc.RunAnalysis(context.Background(), brandID, promptIDs)
	if err != nil {
		log.Printf("Scheduled analysis failed for brand %d: %v", brandID, err)
	} else {
		log.Printf("✅ Scheduled analysis completed for brand %d", brandID)
	}
}
