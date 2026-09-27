// Package versionchecker coordinates checks and upgrades across version sources.
package versionchecker

import "context"

// Drift is one pin whose current version differs from the latest available.
type Drift struct {
	Name    string
	Current string
	Latest  string
}

// Checker reports version drift for one kind of dependency.
type Checker interface {
	Check(context.Context) ([]Drift, error)
}

// Upgrader applies a reported drift. It must reject the change if the pin no
// longer equals Current, then replace it with Latest.
type Upgrader interface {
	Fix(context.Context, Drift) error
}

// Source pairs a checker and upgrader for one kind of dependency.
type Source interface {
	Name() string
	Checker
	Upgrader
}
