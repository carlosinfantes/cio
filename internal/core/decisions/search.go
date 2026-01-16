// Package decisions handles decision storage and management.
package decisions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/carlosinfantes/cto-advisory-board/internal/types"
)

// SearchResult represents a search match with relevance score.
type SearchResult struct {
	Document types.DRFDocument
	Score    int // Higher is more relevant
}

// Search finds DRF documents matching the query across title, intent, tags, and synthesis.
func Search(query string) ([]types.DRFDocument, error) {
	allDocs, err := ListDRFDocuments(nil)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	queryWords := strings.Fields(query)

	var results []SearchResult

	for _, doc := range allDocs {
		score := calculateScore(doc, queryWords)
		if score > 0 {
			results = append(results, SearchResult{
				Document: doc,
				Score:    score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Extract documents
	docs := make([]types.DRFDocument, len(results))
	for i, r := range results {
		docs[i] = r.Document
	}

	return docs, nil
}

// calculateScore determines how well a DRF document matches the query.
func calculateScore(doc types.DRFDocument, queryWords []string) int {
	score := 0
	titleLower := strings.ToLower(doc.Decision.Title)
	intentLower := strings.ToLower(doc.Decision.Intent)
	synthesisLower := strings.ToLower(doc.Synthesis.Decision)

	for _, word := range queryWords {
		// Exact match in title (highest weight)
		if strings.Contains(titleLower, word) {
			score += 10
		}

		// Match in intent (high weight)
		if strings.Contains(intentLower, word) {
			score += 8
		}

		// Match in tags (high weight)
		for _, tag := range doc.Meta.Tags {
			if strings.Contains(strings.ToLower(tag), word) {
				score += 6
			}
		}

		// Match in synthesis (medium weight)
		if strings.Contains(synthesisLower, word) {
			score += 3
		}

		// Match in domain (low weight)
		if strings.Contains(strings.ToLower(doc.Decision.Domain), word) {
			score += 2
		}
	}

	return score
}

// GetRelevantDecisions returns the most relevant past DRF documents for a question.
// Returns up to `limit` documents, sorted by relevance.
func GetRelevantDecisions(question string, limit int) ([]types.DRFDocument, error) {
	if limit <= 0 {
		limit = 5
	}

	results, err := Search(question)
	if err != nil {
		return nil, err
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// FormatDecisionContext creates a text summary of past decisions for LLM context.
func FormatDecisionContext(docs []types.DRFDocument) string {
	if len(docs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Relevant past decisions from this project:\n\n")

	for i, d := range docs {
		sb.WriteString(fmt.Sprintf("%d. %s [%s]\n", i+1, truncateText(d.Decision.Title, 80), d.Meta.Status))
		if d.Synthesis.Decision != "" {
			sb.WriteString(fmt.Sprintf("   → %s\n", truncateText(d.Synthesis.Decision, 120)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// truncateText shortens text to maxLen, adding "..." if truncated.
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
