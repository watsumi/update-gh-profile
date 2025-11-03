package readme

import (
	"fmt"
	"strings"
	"time"
)

// FormatTimestamp formats a timestamp
//
// Preconditions:
// - t is a valid time.Time
// - format is a valid format string (optional, default is RFC3339)
//
// Postconditions:
// - Returns formatted timestamp string
//
// Invariants:
// - Timezone information is included
func FormatTimestamp(t time.Time, format string) string {
	if format == "" {
		// Default format: RFC3339 (e.g., 2024-01-15T10:30:00Z09:00)
		return t.Format(time.RFC3339)
	}

	return t.Format(format)
}

// FormatTimestampWithTimezone formats a timestamp with specified timezone
//
// Preconditions:
// - t is a valid time.Time
// - timezone is a valid timezone name (e.g., "Asia/Tokyo", "UTC")
//
// Postconditions:
// - Returns formatted timestamp string in specified timezone
//
// Invariants:
// - Timezone information is correctly applied
func FormatTimestampWithTimezone(t time.Time, timezone string) (string, error) {
	// Load timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("failed to load timezone: %w", err)
	}

	// Convert to specified timezone
	localTime := t.In(loc)

	// Return in RFC3339 format
	return localTime.Format(time.RFC3339), nil
}

// AddUpdateTimestamp adds update timestamp to specified section in README.md
//
// Preconditions:
// - readmePath is a valid README.md file path
// - sectionTag is the tag name of the section to update
// - timestamp is a valid time.Time (optional, default is current time)
//
// Postconditions:
// - Update timestamp is added to specified section in README.md
// - Added after existing content
//
// Invariants:
// - Timestamp is displayed in a clear format
func AddUpdateTimestamp(readmePath, sectionTag string, timestamp time.Time, timezone string) error {
	// Normalize section tags
	startTag, endTag := NormalizeTags(sectionTag)

	// Read README file
	content, err := ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("failed to read README file: %w", err)
	}

	// Get existing section content (ignore errors and treat as empty string)
	sectionContent, _ := FindSection(content, startTag, endTag)

	// Format timestamp
	var timestampStr string
	if timezone != "" {
		timestampStr, err = FormatTimestampWithTimezone(timestamp, timezone)
		if err != nil {
			return fmt.Errorf("failed to format timestamp: %w", err)
		}
	} else {
		timestampStr = FormatTimestamp(timestamp, "")
	}

	// Generate timestamp markdown
	timestampMarkdown := fmt.Sprintf("\n\n*Last updated: %s*", timestampStr)

	// Create new content (existing content + timestamp)
	newContent := sectionContent
	if sectionContent != "" {
		newContent += timestampMarkdown
	} else {
		newContent = strings.TrimPrefix(timestampMarkdown, "\n\n")
	}

	// Update section (tags are automatically added if they don't exist)
	err = UpdateSection(readmePath, startTag, endTag, newContent)
	if err != nil {
		return fmt.Errorf("failed to update section: %w", err)
	}

	return nil
}

// GenerateTimestampMarkdown generates Markdown format for timestamp
//
// Preconditions:
// - timestamp is a valid time.Time
//
// Postconditions:
// - Returns timestamp string in Markdown format
//
// Invariants:
// - Format is consistent
func GenerateTimestampMarkdown(timestamp time.Time, timezone string) (string, error) {
	var timestampStr string
	var err error

	if timezone != "" {
		timestampStr, err = FormatTimestampWithTimezone(timestamp, timezone)
		if err != nil {
			return "", fmt.Errorf("failed to format timestamp: %w", err)
		}
	} else {
		timestampStr = FormatTimestamp(timestamp, "")
	}

	return fmt.Sprintf("*Last updated: %s*", timestampStr), nil
}

// GetCurrentTimestamp gets current timestamp
//
// Preconditions:
// - None
//
// Postconditions:
// - Returns current time as time.Time
//
// Invariants:
// - Returns in UTC timezone
func GetCurrentTimestamp() time.Time {
	return time.Now().UTC()
}
