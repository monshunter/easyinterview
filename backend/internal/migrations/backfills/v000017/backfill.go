package v000017

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/monshunter/easyinterview/backend/internal/migrations"
)

const (
	Name        = "practice_plan_round_identity"
	batchSize   = 500
	maxSequence = 1<<31 - 1
	zeroUUID    = "00000000-0000-0000-0000-000000000000"
	selectBatch = `
select p.id::text, p.time_budget_minutes, j.summary
from practice_plans p
join target_jobs j
  on j.id = p.target_job_id
 and j.user_id = p.user_id
 and j.resume_id = p.resume_id
 and j.deleted_at is null
where p.round_id is null
  and p.round_sequence is null
  and p.id > $1::uuid
order by p.id
limit $2
`
	applyIdentity = `
update practice_plans
set round_id = $2,
    round_sequence = $3,
    updated_at = now()
where id = $1::uuid
  and round_id is null
  and round_sequence is null
`
)

var validRoundTypes = map[string]struct{}{
	"hr":               {},
	"technical":        {},
	"manager":          {},
	"cross_functional": {},
	"culture":          {},
	"final":            {},
	"other":            {},
}

type candidate struct {
	id                string
	timeBudgetMinutes int
	summary           []byte
}

type targetJobSummary struct {
	InterviewRounds []interviewRound `json:"interviewRounds"`
}

type interviewRound struct {
	Sequence        int    `json:"sequence"`
	Type            string `json:"type"`
	DurationMinutes int    `json:"durationMinutes"`
}

func init() {
	migrations.RegisterBackfill(Name, Run)
}

// Run fills legacy plan round identity only when persisted TargetJob rounds
// contain exactly one duration match. Ambiguous or malformed summaries remain
// untouched so later read models never mistake guessed data for an audited fact.
func Run(ctx context.Context, db *sql.DB, mode migrations.BackfillMode) error {
	if db == nil {
		return fmt.Errorf("practice plan round identity backfill requires a database")
	}
	if mode != migrations.BackfillModeDryRun && mode != migrations.BackfillModeApply {
		return fmt.Errorf("unsupported backfill mode %q", mode)
	}

	cursor := zeroUUID
	for {
		batch, err := loadBatch(ctx, db, cursor)
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}

		for _, item := range batch {
			roundID, sequence, ok := uniqueRoundIdentity(item.summary, item.timeBudgetMinutes)
			if !ok || mode == migrations.BackfillModeDryRun {
				continue
			}
			if _, err := db.ExecContext(ctx, applyIdentity, item.id, roundID, sequence); err != nil {
				return fmt.Errorf("backfill practice plan %s: %w", item.id, err)
			}
		}

		cursor = batch[len(batch)-1].id
		if len(batch) < batchSize {
			return nil
		}
	}
}

func loadBatch(ctx context.Context, db *sql.DB, cursor string) ([]candidate, error) {
	rows, err := db.QueryContext(ctx, selectBatch, cursor, batchSize)
	if err != nil {
		return nil, fmt.Errorf("load legacy practice plans: %w", err)
	}
	defer rows.Close()

	batch := make([]candidate, 0, batchSize)
	for rows.Next() {
		var item candidate
		if err := rows.Scan(&item.id, &item.timeBudgetMinutes, &item.summary); err != nil {
			return nil, fmt.Errorf("scan legacy practice plan: %w", err)
		}
		batch = append(batch, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate legacy practice plans: %w", err)
	}
	return batch, nil
}

func uniqueRoundIdentity(rawSummary []byte, timeBudgetMinutes int) (string, int, bool) {
	var summary targetJobSummary
	if err := json.Unmarshal(rawSummary, &summary); err != nil {
		return "", 0, false
	}

	matches := make([]interviewRound, 0, 1)
	for _, round := range summary.InterviewRounds {
		if round.DurationMinutes == timeBudgetMinutes {
			matches = append(matches, round)
		}
	}
	if len(matches) != 1 {
		return "", 0, false
	}

	match := matches[0]
	roundType := strings.TrimSpace(match.Type)
	if match.Sequence <= 0 || match.Sequence > maxSequence {
		return "", 0, false
	}
	if _, ok := validRoundTypes[roundType]; !ok {
		return "", 0, false
	}
	return fmt.Sprintf("round-%d-%s", match.Sequence, roundType), match.Sequence, true
}
