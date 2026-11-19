package main

import (
	"flag"
	"fmt"
	"io"
	"sort"

	"github.com/use-assay/assay/internal/mechanics"
)

// writeChecks prints every check's ID and Describe() text from the default engine (issue #403).
//
// The check set is bound into the preimage under v2, v3, and v4 encodings
// (attest.PreimageVersionCheckSet and later). Printing the set lets users and
// verifiers see the expected check set without having to extract it from a report.
func writeChecks(w io.Writer, checks []mechanics.Check) error {
	for _, c := range checks {
		if _, err := fmt.Fprintf(w, "%s\t%s\n", c.ID(), c.Describe()); err != nil {
			return fmt.Errorf("write checks: %w", err)
		}
	}
	return nil
}

// defaultChecks returns the default engine's checks, sorted by check ID.
// Sorting ensures the iteration order matches mechanics.Engine.CheckIDs() and
// the order checks appear in v2/v3/v4 preimage encodings.
func defaultChecks() []mechanics.Check {
	eng := mechanics.NewEngine()
	checks := append([]mechanics.Check(nil), eng.Checks...)
	sort.Slice(checks, func(i, j int) bool {
		return checks[i].ID() < checks[j].ID()
	})
	return checks
}

// runChecks implements `assay checks`.
func runChecks(args []string, d commandDeps) error {
	fs := flag.NewFlagSet("checks", flag.ContinueOnError)
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return usagef("checks takes no arguments (got %q)", fs.Arg(0))
	}
	return writeChecks(d.stdout, defaultChecks())
}
