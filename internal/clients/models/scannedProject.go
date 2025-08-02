package models

import "time"

type ScannedPackage struct {
	ServiceName      string          `json:"-"`
	Name             string          `json:"-"`
	ProjectName      string          `json:"-"`
	Summary          string          `json:"summary"`
	Description      string          `json:"description"`
	Severity         string          `json:"severity"`
	GithubReviewedAt time.Time       `json:"github_reviewed_at"`
	Vulnerabilities  []Vulnerability `json:"vulnerabilities"`
	RiskScore        int             `json:"-"`
}
