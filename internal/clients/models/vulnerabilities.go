package models

type Vulnerability struct {
	Name                   string  `json:"name"`
	CurrentVersion         string  `json:"-"`
	Package                Package `json:"package"`
	VulnerableVersionRange string  `json:"vulnerable_version_range"`
	FirstPatchedVersion    string  `json:"first_patched_version"`
}
