package mechanics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/use-assay/assay/internal/horizon"
	"github.com/use-assay/assay/internal/mechanics"
)

func validSubject() *mechanics.Subject {
	return &mechanics.Subject{
		Asset: mechanics.Asset{Code: "USDC", Issuer: "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN"},
		Stat: &horizon.AssetStat{
			AssetCode:   "USDC",
			AssetIssuer: "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN",
			Flags: horizon.Flags{
				AuthRequired:  false,
				AuthRevocable: false,
			},
		},
		StatFetchedAt: time.Now(),
		Issuer: &horizon.Account{
			AccountID: "GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN",
			Flags: horizon.Flags{
				AuthRequired:  false,
				AuthRevocable: false,
			},
		},
		IssuerFetchedAt: time.Now(),
		ScannedAt:       time.Now(),
	}
}

func TestPartialCheckSetRequiresExplicitOptIn(t *testing.T) {
	ctx := context.Background()
	sub := validSubject()

	cases := []struct {
		name   string
		checks []mechanics.Check
	}{
		{
			name:   "empty check set",
			checks: []mechanics.Check{},
		},
		{
			name: "single capability check",
			checks: []mechanics.Check{
				mechanics.CapabilityCheck{},
			},
		},
		{
			name: "missing reputation check",
			checks: []mechanics.Check{
				mechanics.CapabilityCheck{},
				mechanics.MutabilityCheck{},
				mechanics.DomainCheck{},
			},
		},
		{
			name: "missing capability check",
			checks: []mechanics.Check{
				mechanics.MutabilityCheck{},
				mechanics.DomainCheck{},
				mechanics.ReputationCheck{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eng := &mechanics.Engine{Checks: tc.checks}
			if !eng.IsPartialCheckSet() {
				t.Fatalf("expected IsPartialCheckSet() to be true for %s", tc.name)
			}

			rep, err := eng.Run(ctx, sub)
			if err == nil {
				t.Fatalf("expected Run to fail for partial check set without AllowPartialCheckSet")
			}
			if !errors.Is(err, mechanics.ErrPartialCheckSet) {
				t.Fatalf("got error %v, want errors.Is(err, mechanics.ErrPartialCheckSet)", err)
			}
			if rep != nil {
				t.Fatalf("expected nil report on refusal, got %+v", rep)
			}
		})
	}
}

func TestPartialCheckSetAllowedWithExplicitOptIn(t *testing.T) {
	ctx := context.Background()
	sub := validSubject()

	t.Run("struct field", func(t *testing.T) {
		eng := &mechanics.Engine{
			Checks: []mechanics.Check{
				mechanics.CapabilityCheck{},
				mechanics.MutabilityCheck{},
			},
			AllowPartialCheckSet: true,
		}

		if !eng.IsPartialCheckSet() {
			t.Fatal("expected IsPartialCheckSet() to be true")
		}

		rep, err := eng.Run(ctx, sub)
		if err != nil {
			t.Fatalf("Run failed despite AllowPartialCheckSet: %v", err)
		}
		if !rep.PartialCheckSet {
			t.Error("rep.PartialCheckSet = false, want true")
		}
		if !rep.AllowPartialCheckSet {
			t.Error("rep.AllowPartialCheckSet = false, want true")
		}
	})

	t.Run("functional option", func(t *testing.T) {
		eng := mechanics.NewEngine(
			mechanics.WithChecks(mechanics.CapabilityCheck{}),
			mechanics.AllowPartialCheckSet(),
		)

		if !eng.AllowPartialCheckSet {
			t.Fatal("expected AllowPartialCheckSet to be true on engine")
		}
		if !eng.IsPartialCheckSet() {
			t.Fatal("expected IsPartialCheckSet() to be true")
		}

		rep, err := eng.Run(ctx, sub)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}
		if !rep.PartialCheckSet {
			t.Error("rep.PartialCheckSet = false, want true")
		}
	})
}

func TestDefaultAndSupersetCheckSetsAreNotPartial(t *testing.T) {
	ctx := context.Background()
	sub := validSubject()

	t.Run("default engine", func(t *testing.T) {
		eng := mechanics.NewEngine()
		if eng.IsPartialCheckSet() {
			t.Fatal("default engine should not be partial")
		}
		rep, err := eng.Run(ctx, sub)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if rep.PartialCheckSet {
			t.Error("rep.PartialCheckSet = true on default engine, want false")
		}
		if rep.AllowPartialCheckSet {
			t.Error("rep.AllowPartialCheckSet = true on default engine, want false")
		}
	})

	t.Run("superset with trustline check", func(t *testing.T) {
		eng := &mechanics.Engine{
			Checks: append(append([]mechanics.Check(nil), mechanics.NewEngine().Checks...), mechanics.TrustlineCheck{}),
		}
		if eng.IsPartialCheckSet() {
			t.Fatal("superset engine should not be partial")
		}
		rep, err := eng.Run(ctx, sub)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if rep.PartialCheckSet {
			t.Error("rep.PartialCheckSet = true on superset engine, want false")
		}
	})
}
