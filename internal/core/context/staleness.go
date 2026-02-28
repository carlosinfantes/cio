// Package context handles loading and managing CRF context entities.
package context

import (
	"github.com/carlosinfantes/cio/internal/types"
)

// DefaultStalenessThreshold is the number of days before context is considered stale.
const DefaultStalenessThreshold = 30

// CheckStaleness checks if the CRF context is outdated.
// Returns a warning if the context is older than the threshold.
// Note: This function is now deprecated in favor of CheckContextStaleness in loader.go
// which works directly with CRFContext.
func CheckStaleness(ctx *types.CRFContext, thresholdDays int) *types.StalenessWarning {
	return CheckContextStaleness(ctx, thresholdDays)
}
