package application

import (
	"time"
)

// Example file with mixed @Service and @RouterService annotations

// ============================================================
// Pure Service (No HTTP)
// ============================================================

// @Service name="email-service"
type EmailService struct {
	// @InjectCfg key="smtp.host"
	SMTPHost string

	// @InjectCfg key="smtp.port", default=587
	SMTPPort int

	// @InjectCfg key="smtp.username"
	SMTPUsername string

	// @InjectCfg key="smtp.password"
	SMTPPassword string

	// @InjectCfg key="email.from", default="noreply@example.com"
	FromEmail string
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	// Send email implementation
	println("Sending email from", s.FromEmail, "to", to, "via", s.SMTPHost)
	return nil
}

// ============================================================
// Another Pure Service
// ============================================================

// @Service name="background-job-service"
type BackgroundJobService struct {
	// @Inject "email-service"
	EmailService *EmailService

	// @InjectCfg key="jobs.max-workers", default=10
	MaxWorkers int

	// @InjectCfg key="jobs.retry-limit", default=3
	RetryLimit int

	// @InjectCfg key="jobs.retry-delay", default="5s"
	RetryDelay time.Duration
}

func (s *BackgroundJobService) ProcessJob(jobID string) error {
	// Process job and send notification
	return s.EmailService.SendEmail("admin@example.com", "Job Complete", "Job "+jobID+" completed")
}

// ============================================================
// HTTP Service (RouterService)
// ============================================================

// @RouterService name="admin-api-service", prefix="/api/admin", middlewares=["recovery", "auth", "admin"]
type AdminAPIService struct {
	// @Inject "background-job-service"
	JobService *BackgroundJobService

	// @Inject service="email-service"
	EmailService *EmailService

	// @InjectCfg key="admin.allow-job-restart", default=true
	AllowJobRestart bool

	// @InjectCfg key="admin.max-jobs-per-page", default=50
	MaxJobsPerPage int
}

// @Route "POST /jobs/{id}/restart"
func (s *AdminAPIService) RestartJob(req *RestartJobRequest) (*JobResponse, error) {
	if !s.AllowJobRestart {
		return nil, nil
	}

	err := s.JobService.ProcessJob(req.JobID)
	if err != nil {
		return nil, err
	}

	return &JobResponse{
		JobID:  req.JobID,
		Status: "restarted",
	}, nil
}

// @Route "GET /jobs"
func (s *AdminAPIService) ListJobs(req *ListJobsRequest) (*JobListResponse, error) {
	// List jobs with pagination
	return &JobListResponse{
		Jobs:     []*JobResponse{},
		PageSize: s.MaxJobsPerPage,
	}, nil
}

// Request/Response DTOs for AdminAPIService
type RestartJobRequest struct {
	JobID string `path:"id"`
}

type ListJobsRequest struct {
	Page int `query:"page"`
}

type JobResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

type JobListResponse struct {
	Jobs     []*JobResponse `json:"jobs"`
	PageSize int            `json:"page_size"`
}
