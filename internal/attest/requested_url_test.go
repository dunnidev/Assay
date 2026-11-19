package attest_test

import (
	"testing"

	"github.com/use-assay/assay/internal/attest"
)

// RequestedURL is audit information added alongside the URL a claim is
// attributed to. It must NOT be part of the evidence_hash preimage: adding it
// has to leave the hash of an otherwise unchanged claim alone, or every
// attestation already on chain stops reproducing. If a future encoding binds
// it, that is a deliberate PreimageVersion bump, not this.
func TestRequestedURLIsOutsideEvidenceHash(t *testing.T) {
	base, err := attest.FromReport(report(nil))
	if err != nil {
		t.Fatalf("FromReport(base): %v", err)
	}

	withRequested := report(nil)
	withRequested.Evidence[0].RequestedURL = "https://circle.com/.well-known/stellar.toml"
	got, err := attest.FromReport(withRequested)
	if err != nil {
		t.Fatalf("FromReport(with RequestedURL): %v", err)
	}

	if base.EvidenceHash != got.EvidenceHash {
		t.Errorf("recording RequestedURL moved evidence_hash: %s != %s; it must stay outside the preimage",
			base.EvidenceHash, got.EvidenceHash)
	}
	if base.Preimage != got.Preimage {
		t.Errorf("recording RequestedURL changed the preimage:\n got: %q\nwant: %q",
			got.Preimage, base.Preimage)
	}
}

// LatestLedger and LedgerSequence are audit metadata added alongside URL and
// RetrievedAt for Horizon evidence (issue #402). They must NOT be part of the
// evidence_hash preimage: adding them has to leave the hash and preimage of an
// otherwise unchanged claim alone, matching RequestedURL.
func TestLedgerSequenceIsOutsideEvidenceHash(t *testing.T) {
	base, err := attest.FromReport(report(nil))
	if err != nil {
		t.Fatalf("FromReport(base): %v", err)
	}

	withLedger := report(nil)
	withLedger.Evidence[0].LatestLedger = 55678123
	withLedger.Evidence[0].LedgerSequence = 55678123
	got, err := attest.FromReport(withLedger)
	if err != nil {
		t.Fatalf("FromReport(with LatestLedger): %v", err)
	}

	if base.EvidenceHash != got.EvidenceHash {
		t.Errorf("recording LatestLedger moved evidence_hash: %s != %s; it must stay outside the preimage",
			base.EvidenceHash, got.EvidenceHash)
	}
	if base.Preimage != got.Preimage {
		t.Errorf("recording LatestLedger changed the preimage:\n got: %q\nwant: %q",
			got.Preimage, base.Preimage)
	}
}
