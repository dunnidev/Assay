package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/use-assay/assay/internal/mechanics"
)

// TestChecksOutputFormat verifies that `assay checks` prints each check's ID
// and Describe() text, separated by a tab.
func TestChecksOutputFormat(t *testing.T) {
	d, out, _, _ := deps(nil, nil)
	if err := runWith([]string{"checks"}, d); err != nil {
		t.Fatalf("runWith(checks): %v", err)
	}

	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	expectedChecks := defaultChecks()

	if len(lines) != len(expectedChecks) {
		t.Fatalf("got %d lines, want %d", len(lines), len(expectedChecks))
	}

	for i, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			t.Fatalf("line %d does not contain exactly two tab-separated fields: %q", i, line)
		}
		gotID, gotDesc := parts[0], parts[1]
		wantCheck := expectedChecks[i]

		if gotID != wantCheck.ID() {
			t.Errorf("line %d ID = %q, want %q", i, gotID, wantCheck.ID())
		}
		if gotDesc != wantCheck.Describe() {
			t.Errorf("line %d description = %q, want %q", i, gotDesc, wantCheck.Describe())
		}
	}
}

// TestChecksMatchesPreimageVectors verifies that the check set printed by
// `assay checks` matches the check set bound into v2 and v3 preimages (#403).
func TestChecksMatchesPreimageVectors(t *testing.T) {
	d, out, _, _ := deps(nil, nil)
	if err := runWith([]string{"checks"}, d); err != nil {
		t.Fatalf("runWith(checks): %v", err)
	}

	var printedIDs []string
	for _, line := range strings.Split(strings.TrimRight(out.String(), "\n"), "\n") {
		parts := strings.Split(line, "\t")
		printedIDs = append(printedIDs, parts[0])
	}

	// 1. Matches mechanics.NewEngine().CheckIDs()
	eng := mechanics.NewEngine()
	wantIDs := eng.CheckIDs()
	if !reflect.DeepEqual(printedIDs, wantIDs) {
		t.Errorf("printed IDs = %v, want engine check IDs = %v", printedIDs, wantIDs)
	}

	// 2. Matches committed v2 and v3 vector preimages
	for _, vectorFile := range []string{
		"../../internal/attest/testdata/vectors/checkset-bound-v2.preimage",
		"../../internal/attest/testdata/vectors/network-bound-pubnet.preimage",
	} {
		data, err := os.ReadFile(filepath.Clean(vectorFile))
		if err != nil {
			t.Fatalf("read vector %s: %v", vectorFile, err)
		}
		var vectorChecks []string
		for _, vline := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(vline, "checks\t") {
				raw := strings.TrimPrefix(vline, "checks\t")
				vectorChecks = strings.Split(raw, ",")
				break
			}
		}
		if len(vectorChecks) == 0 {
			t.Fatalf("vector %s did not contain checks line", vectorFile)
		}
		if !reflect.DeepEqual(printedIDs, vectorChecks) {
			t.Errorf("printed IDs = %v do not match vector %s checks = %v",
				printedIDs, vectorFile, vectorChecks)
		}
	}
}

// TestChecksTakesNoArguments verifies that arguments to `checks` are rejected
// with exitUsage.
func TestChecksTakesNoArguments(t *testing.T) {
	d, _, _, _ := deps(nil, nil)
	err := runWith([]string{"checks", "extra"}, d)
	if err == nil {
		t.Fatal("checks with an argument should fail")
	}
	if got := exitCodeForError(err); got != exitUsage {
		t.Errorf("checks extra: exit = %d, want %d (usage)", got, exitUsage)
	}
}
