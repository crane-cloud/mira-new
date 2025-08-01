package common

import "fmt"

// NATS Subject Constants

const (
	// Core subjects
	SUBJECT_BUILD_REQUEST = "mira.build.requests"

	// Subject patterns (use with fmt.Sprintf)
	SUBJECT_BUILD_STATUS_PATTERN = "mira.status.%s" // %s = buildID
	SUBJECT_BUILD_LOGS_PATTERN   = "mira.logs.%s"   // %s = buildID
)

// Subject builders for dynamic subjects

// BuildStatusSubject returns the subject for build status updates for a specific build
func BuildStatusSubject(buildID string) string {
	return fmt.Sprintf(SUBJECT_BUILD_STATUS_PATTERN, buildID)
}

// BuildLogsSubject returns the subject for build logs for a specific build
func BuildLogsSubject(buildID string) string {
	return fmt.Sprintf(SUBJECT_BUILD_LOGS_PATTERN, buildID)
}

// NATSSubjects contains all NATS subject information for documentation
type NATSSubjects struct {
	BuildRequests string
	BuildStatus   string
	BuildLogs     string
}

// GetSubjectsDocumentation returns documentation about all NATS subjects
func GetSubjectsDocumentation() NATSSubjects {
	return NATSSubjects{
		BuildRequests: SUBJECT_BUILD_REQUEST + " - Queue for containerization build requests",
		BuildStatus:   SUBJECT_BUILD_STATUS_PATTERN + " - Build status updates (running, completed, failed)",
		BuildLogs:     SUBJECT_BUILD_LOGS_PATTERN + " - Real-time build logs stream",
	}
}

// ValidateSubject checks if a subject follows the expected patterns
func ValidateSubject(subject string) bool {
	switch {
	case subject == SUBJECT_BUILD_REQUEST:
		return true
	case len(subject) > len("mira.status.") && subject[:len("mira.status.")] == "mira.status.":
		return true
	case len(subject) > len("mira.logs.") && subject[:len("mira.logs.")] == "mira.logs.":
		return true
	default:
		return false
	}
}
