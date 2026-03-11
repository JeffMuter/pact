package buckets

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateDurationMinutes_OnlyDays(t *testing.T) {
	result := calculateDurationMinutes(
		sql.NullInt64{Int64: 2, Valid: true},
		sql.NullInt64{Valid: false},
		sql.NullInt64{Valid: false},
	)
	// 2 days * 1440 minutes/day = 2880 minutes
	assert.Equal(t, int64(2880), result)
}

func TestCalculateDurationMinutes_OnlyHours(t *testing.T) {
	result := calculateDurationMinutes(
		sql.NullInt64{Valid: false},
		sql.NullInt64{Int64: 3, Valid: true},
		sql.NullInt64{Valid: false},
	)
	// 3 hours * 60 minutes/hour = 180 minutes
	assert.Equal(t, int64(180), result)
}

func TestCalculateDurationMinutes_OnlyMinutes(t *testing.T) {
	result := calculateDurationMinutes(
		sql.NullInt64{Valid: false},
		sql.NullInt64{Valid: false},
		sql.NullInt64{Int64: 45, Valid: true},
	)
	// 45 minutes = 45 minutes
	assert.Equal(t, int64(45), result)
}

func TestCalculateDurationMinutes_Combined(t *testing.T) {
	result := calculateDurationMinutes(
		sql.NullInt64{Int64: 1, Valid: true},
		sql.NullInt64{Int64: 2, Valid: true},
		sql.NullInt64{Int64: 30, Valid: true},
	)
	// 1 day (1440) + 2 hours (120) + 30 minutes = 1590 minutes
	assert.Equal(t, int64(1590), result)
}

func TestCalculateDurationMinutes_AllNull(t *testing.T) {
	result := calculateDurationMinutes(
		sql.NullInt64{Valid: false},
		sql.NullInt64{Valid: false},
		sql.NullInt64{Valid: false},
	)
	// Default to 1440 minutes (24 hours)
	assert.Equal(t, int64(1440), result)
}
