package readme

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// UpdateSection replaces the specified section in README.md with new content
// Automatically appends tags if they don't exist
//
// Preconditions:
// - readmePath is a valid README.md file path
// - startTag and endTag are valid comment tag strings
// - newContent is the new content to use for replacement
//
// Postconditions:
// - Content between startTag and endTag in README.md is replaced with newContent
// - Tags are appended to the end of the file if they don't exist
// - Existing content (except tags) is preserved
//
// Invariants:
// - Returns error if file doesn't exist
func UpdateSection(readmePath, startTag, endTag, newContent string) error {
	// Read README file
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("failed to read README file: %w", err)
	}

	readmeContent := string(content)

	// Replace content between tags
	updatedContent, err := ReplaceSectionOrAppend(readmeContent, startTag, endTag, newContent)
	if err != nil {
		return fmt.Errorf("failed to replace section: %w", err)
	}

	// Write updated content to file
	err = os.WriteFile(readmePath, []byte(updatedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write README file: %w", err)
	}

	return nil
}

// ReplaceSectionOrAppend replaces the specified section (between startTag and endTag) in text with new content
// Automatically appends to end of file if tags don't exist
//
// Preconditions:
// - content is the content of a loaded file
// - startTag and endTag are valid comment tag strings
//
// Postconditions:
// - Content between startTag and endTag is replaced with newContent
// - Tags are appended to the end of the file if they don't exist
// - Tags themselves are preserved
//
// Invariants:
// - Existing content (except tags) is preserved
func ReplaceSectionOrAppend(content, startTag, endTag, newContent string) (string, error) {
	// Check if startTag and endTag exist
	startIndex := strings.Index(content, startTag)
	endIndex := strings.Index(content, endTag)

	// Append to end of file if tags don't exist
	if startIndex == -1 || endIndex == -1 {
		// Add newline if end of file doesn't have one
		result := content
		if !strings.HasSuffix(result, "\n") && !strings.HasSuffix(result, "\n\n") {
			result += "\n"
		}
		// Add tags and content
		result += "\n" + startTag + "\n"
		if newContent != "" {
			result += newContent + "\n"
		}
		result += endTag + "\n"
		return result, nil
	}

	// Error if endTag position is before startTag
	if endIndex < startIndex {
		return "", fmt.Errorf("end tag is before start tag")
	}

	// Use existing ReplaceSection logic
	return ReplaceSection(content, startTag, endTag, newContent)
}

// ReplaceSection replaces the specified section (between startTag and endTag) in text with new content
//
// Preconditions:
// - content is the content of a loaded file
// - startTag and endTag are valid comment tag strings
//
// Postconditions:
// - Content between startTag and endTag is replaced with newContent
// - Tags themselves are preserved
//
// Invariants:
// - Returns error if tags are not found
// - Existing content (except tags) is preserved
func ReplaceSection(content, startTag, endTag, newContent string) (string, error) {
	// Check if startTag and endTag exist
	startIndex := strings.Index(content, startTag)
	if startIndex == -1 {
		return "", fmt.Errorf("start tag not found: %s", startTag)
	}

	endIndex := strings.Index(content, endTag)
	if endIndex == -1 {
		return "", fmt.Errorf("end tag not found: %s", endTag)
	}

	// Error if endTag position is before startTag
	if endIndex < startIndex {
		return "", fmt.Errorf("end tag is before start tag")
	}

	// Get end position of startTag (up to tag end)
	startTagEnd := startIndex + len(startTag)

	// Start position of endTag (tag start position)

	// Replace: [end of start tag] + [new content] + [end tag]
	before := content[:startTagEnd]
	after := content[endIndex:]

	// Adjust newlines
	result := before

	// Preserve newlines even if new content is empty
	if newContent == "" {
		// Add newline if tag doesn't have one after it
		if !strings.HasSuffix(before, "\n") {
			result += "\n"
		}
	} else {
		// Add newline if tag doesn't have one after it
		if !strings.HasSuffix(before, "\n") {
			result += "\n"
		}
		result += newContent
		// Add newline before endTag if needed
		if !strings.HasPrefix(after, "\n") && !strings.HasSuffix(newContent, "\n") {
			result += "\n"
		}
	}

	result += after

	return result, nil
}

// FindSection extracts content between specified tags
//
// Preconditions:
// - content is the content of a loaded file
// - startTag and endTag are valid comment tag strings
//
// Postconditions:
// - Returns content between startTag and endTag (tags themselves not included)
// - Returns error if tags are not found
//
// Invariants:
// - Tag order is verified as correct
func FindSection(content, startTag, endTag string) (string, error) {
	startIndex := strings.Index(content, startTag)
	if startIndex == -1 {
		return "", fmt.Errorf("start tag not found: %s", startTag)
	}

	endIndex := strings.Index(content, endTag)
	if endIndex == -1 {
		return "", fmt.Errorf("end tag not found: %s", endTag)
	}

	if endIndex < startIndex {
		return "", fmt.Errorf("end tag is before start tag")
	}

	// Extract from end of start tag to start of end tag
	startContentIndex := startIndex + len(startTag)
	sectionContent := content[startContentIndex:endIndex]

	// Trim whitespace and newlines from front and back
	sectionContent = strings.TrimSpace(sectionContent)

	return sectionContent, nil
}

// ValidateTags validates that specified tags exist correctly in README.md
//
// Preconditions:
// - readmePath is a valid README.md file path
// - startTag and endTag are valid comment tag strings
//
// Postconditions:
// - Returns nil if tags exist and order is correct
// - Returns error if tags are not found or order is incorrect
//
// Invariants:
// - Assumes file can be read
func ValidateTags(readmePath, startTag, endTag string) error {
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return fmt.Errorf("failed to read README file: %w", err)
	}

	readmeContent := string(content)

	// Check tag existence
	startIndex := strings.Index(readmeContent, startTag)
	if startIndex == -1 {
		return fmt.Errorf("start tag not found: %s", startTag)
	}

	endIndex := strings.Index(readmeContent, endTag)
	if endIndex == -1 {
		return fmt.Errorf("end tag not found: %s", endTag)
	}

	// Verify order
	if endIndex < startIndex {
		return fmt.Errorf("end tag is before start tag")
	}

	return nil
}

// ReadFile reads README.md file
//
// Preconditions:
// - filePath is a valid file path
//
// Postconditions:
// - Returns file content as a string
// - Returns error if file doesn't exist
//
// Invariants:
// - Assumes file is UTF-8 encoded
func ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// NormalizeTags normalizes comment tags (unifies case, trims whitespace)
//
// Preconditions:
// - tag is a comment tag string (e.g., "START_LANGUAGE_STATS")
//
// Postconditions:
// - Returns HTML comment format tags (e.g., "<!-- START_LANGUAGE_STATS -->")
//
// Invariants:
// - Returned tags are valid HTML comment format
func NormalizeTags(tagName string) (startTag, endTag string) {
	// If already in HTML comment format
	if strings.HasPrefix(tagName, "<!--") {
		// Extract tag name part (remove "<!-- " and " -->")
		tagContent := strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(tagName, "-->"), "<!--"))
		if strings.HasPrefix(tagContent, "START") {
			// If START tag, generate END tag
			endContent := strings.TrimSpace(strings.Replace(tagContent, "START", "END", 1))
			return tagName, fmt.Sprintf("<!-- %s -->", endContent)
		}
		// Otherwise use as is
		return tagName, fmt.Sprintf("<!-- END%s -->", strings.TrimPrefix(tagContent, "START"))
	}

	// Generate HTML comment format from tag name (convert to uppercase)
	startTag = fmt.Sprintf("<!-- START_%s -->", strings.ToUpper(tagName))
	endTag = fmt.Sprintf("<!-- END_%s -->", strings.ToUpper(tagName))

	return startTag, endTag
}

// EscapeRegexSpecialChars escapes regex special characters (internal use)
func escapeRegexSpecialChars(s string) string {
	re := regexp.MustCompile(`[.*+?^${}()|[\]\\]`)
	return re.ReplaceAllStringFunc(s, func(matched string) string {
		return "\\" + matched
	})
}
