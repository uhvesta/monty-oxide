package versionchecker

import (
	"context"
	"fmt"
)

// Report identifies which source found a drift.
type Report struct {
	Source string
	Drift  Drift
}

// Runner calls each registered source in order.
type Runner struct {
	Sources []Source
}

// Check reports drift without changing pins.
func (r Runner) Check(ctx context.Context) ([]Report, error) {
	return r.run(ctx, false)
}

// Fix checks for drift, then applies the latest version through each source.
func (r Runner) Fix(ctx context.Context) ([]Report, error) {
	return r.run(ctx, true)
}

func (r Runner) run(ctx context.Context, fix bool) ([]Report, error) {
	if len(r.Sources) == 0 {
		return nil, fmt.Errorf("no version sources registered")
	}
	var reports []Report
	for _, source := range r.Sources {
		if err := ctx.Err(); err != nil {
			return reports, err
		}
		drifts, err := source.Check(ctx)
		if err != nil {
			return reports, fmt.Errorf("%s: check: %w", source.Name(), err)
		}
		for _, drift := range drifts {
			if drift.Current == drift.Latest {
				continue
			}
			if fix {
				err := source.Fix(ctx, drift)
				if err != nil {
					return reports, fmt.Errorf("%s/%s: fix: %w", source.Name(), drift.Name, err)
				}
			}
			reports = append(reports, Report{Source: source.Name(), Drift: drift})
		}
		if fix && len(drifts) > 0 {
			if finisher, ok := source.(interface{ FinishFix(context.Context) error }); ok {
				if err := finisher.FinishFix(ctx); err != nil {
					return reports, fmt.Errorf("%s: finish fix: %w", source.Name(), err)
				}
			}
		}
	}
	return reports, nil
}
