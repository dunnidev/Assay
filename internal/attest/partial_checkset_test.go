package attest_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/use-assay/assay/internal/attest"
	"github.com/use-assay/assay/internal/horizon"
	"github.com/use-assay/assay/internal/mechanics"
	"github.com/use-assay/assay/internal/stellarexpert"
)

func validAttestableSubject() *mechanics.Subject {
	now := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	const issuer = "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"
	const domain = "example.com"
	return &mechanics.Subject{
		Asset: mechanics.Asset{Code: "USDC", Issuer: issuer},
		Stat: &horizon.AssetStat{
			AssetCode:   "USDC",
			AssetIssuer: issuer,
			Flags: horizon.Flags{
				AuthRequired:  false,
				AuthRevocable: false,
			},
		},
		StatFetchedAt: now,
		Issuer: &horizon.Account{
			AccountID:  issuer,
			HomeDomain: domain,
			Flags: horizon.Flags{
				AuthRequired:  false,
				AuthRevocable: false,
			},
		},
		IssuerFetchedAt:      now,
		DirectoryURL:         "https://api.stellar.expert/explorer/directory/" + issuer,
		DirectoryAttemptedAt: now,
		BlockedURL:           "https://api.stellar.expert/explorer/directory/blocked-domains/" + domain,
		Blocked:              &stellarexpert.BlockedDomain{Domain: domain, Blocked: false},
		BlockedFetchedAt:     now,
		BlockedAttemptedAt:   now,
		ScannedAt:            now,
	}
}

// TestFromReportRefusesPartialCheckSet verifies acceptance criterion:
// Reports produced by a partial engine are refused by FromReport (#404).
func TestFromReportRefusesPartialCheckSet(t *testing.T) {
	ctx := context.Background()
	sub := validAttestableSubject()

	// An engine with only CapabilityCheck (missing mutability, domain, reputation)
	eng := &mechanics.Engine{
		Checks: []mechanics.Check{
			mechanics.CapabilityCheck{},
		},
		AllowPartialCheckSet: true,
	}

	rep, err := eng.Run(ctx, sub)
	if err != nil {
		t.Fatalf("eng.Run: %v", err)
	}
	if !rep.PartialCheckSet {
		t.Fatal("expected rep.PartialCheckSet to be true")
	}

	// FromReport must refuse it
	_, err = attest.FromReport(rep)
	if err == nil {
		t.Fatal("expected FromReport to refuse report with PartialCheckSet = true")
	}
	if !errors.Is(err, attest.ErrPartialCheckSet) {
		t.Fatalf("got err %v, want errors.Is(err, attest.ErrPartialCheckSet)", err)
	}

	// FromReportAs under historical versions must also refuse it
	for _, ver := range []string{"v1", "v2", "v3", "v4"} {
		_, err := attest.FromReportAs(rep, ver)
		if err == nil {
			t.Fatalf("FromReportAs(%s) allowed partial check set report", ver)
		}
		if !errors.Is(err, attest.ErrPartialCheckSet) {
			t.Fatalf("FromReportAs(%s) err = %v, want ErrPartialCheckSet", ver, err)
		}
	}
}

// TestFromReportAcceptsCompleteReport verifies that a complete scan from the
// default engine is accepted by FromReport.
func TestFromReportAcceptsCompleteReport(t *testing.T) {
	ctx := context.Background()
	sub := validAttestableSubject()

	eng := mechanics.NewEngine()
	rep, err := eng.Run(ctx, sub)
	if err != nil {
		t.Fatalf("eng.Run: %v", err)
	}
	if rep.PartialCheckSet {
		t.Fatal("default engine report should not have PartialCheckSet = true")
	}

	params, err := attest.FromReport(rep)
	if err != nil {
		t.Fatalf("FromReport: %v", err)
	}
	if params.EvidenceHash == "" {
		t.Fatal("expected non-empty EvidenceHash")
	}
}
