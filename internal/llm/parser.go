package llm

import (
	"strings"
	"github.com/ozankasikci/gitai/internal/logger"
)

func parseResponse(response string) []CommitSuggestion {
	logger.Debugf("\n=== Starting to parse response ===\nFull response text:\n%s\n", response)
	var suggestions []CommitSuggestion
	lines := strings.Split(response, "\n")
	logger.Debugf("Split response into %d lines", len(lines))

	var currentSuggestion *CommitSuggestion
	for i, line := range lines {
		line = strings.TrimSpace(line)
		logger.Debugf("Line %d: '%s'", i+1, line)
		
		if line == "" {
			logger.Debugf("Skipping empty line")
			continue
		}

		// Check if this is a new suggestion line (matches "1.", "1 -", or "1. -")
		if len(line) > 2 && line[0] >= '1' && line[0] <= '9' {
			// If we have a previous suggestion, add it
			if currentSuggestion != nil {
				logger.Debugf("Adding previous suggestion: %+v", *currentSuggestion)
				suggestions = append(suggestions, *currentSuggestion)
			}

			// Start new suggestion
			currentSuggestion = &CommitSuggestion{}
			
			// Extract message part
			if strings.Contains(line, " - ") {
				parts := strings.SplitN(line, " - ", 2)
				if len(parts) == 2 {
					currentSuggestion.Message = strings.TrimSpace(parts[1])
				}
			} else if strings.Contains(line, ". ") {
				parts := strings.SplitN(line, ". ", 2)
				if len(parts) == 2 {
					currentSuggestion.Message = strings.TrimSpace(parts[1])
				}
			}
			
			logger.Debugf("Created new suggestion with message: %s", currentSuggestion.Message)
		} else if currentSuggestion != nil {
			lowercaseLine := strings.ToLower(line)
			if strings.HasPrefix(lowercaseLine, "explanation:") {
				 explanation := strings.TrimSpace(strings.TrimPrefix(line, "Explanation:"))
				 logger.Debugf("Adding explanation to current suggestion: %s", explanation)
				 currentSuggestion.Explanation = explanation
			}
		}
	}

	// Add the last suggestion
	if currentSuggestion != nil {
		logger.Debugf("Adding final suggestion: %+v", *currentSuggestion)
		suggestions = append(suggestions, *currentSuggestion)
	}

	logger.Debugf("\n=== Parsing complete ===\nFound %d suggestions", len(suggestions))
	for i, s := range suggestions {
		logger.Debugf("Suggestion %d: Message='%s', Explanation='%s'", i+1, s.Message, s.Explanation)
	}
	
	return suggestions
} 