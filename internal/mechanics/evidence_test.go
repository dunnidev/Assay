package mechanics_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/use-assay/assay/internal/mechanics"
)

func TestEvidenceLedgerSequenceJSONRoundTrip(t *testing.T) {
	ev := mechanics.Evidence{
		Source:         "horizon",
		URL:            "https://horizon.stellar.org/accounts/GA5ZSEJYB37JRC5AVCIA5MOP4RHTM335X2KGX3IHOJAPP5RE34K4KZVN",
		Claim:          "auth_required=false auth_revocable=false auth_immutable=false auth_clawback_enabled=false",
		LatestLedger:   55678123,
		LedgerSequence: 55678123,
	}

	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"latest_ledger":55678123`) {
		t.Errorf("expected latest_ledger in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"ledger_sequence":55678123`) {
		t.Errorf("expected ledger_sequence in JSON: %s", jsonStr)
	}

	var unmarshaled mechanics.Evidence
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if unmarshaled.LatestLedger != 55678123 || unmarshaled.LedgerSequence != 55678123 {
		t.Errorf("round trip mismatch: got LatestLedger=%d, LedgerSequence=%d",
			unmarshaled.LatestLedger, unmarshaled.LedgerSequence)
	}
}

func TestEvidenceUnmarshalJSONCompatibility(t *testing.T) {
	t.Run("latest_ledger only", func(t *testing.T) {
		raw := `{"source":"horizon","url":"https://h","claim":"c","latest_ledger":123456}`
		var ev mechanics.Evidence
		if err := json.Unmarshal([]byte(raw), &ev); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if ev.LatestLedger != 123456 || ev.LedgerSequence != 123456 {
			t.Errorf("got LatestLedger=%d, LedgerSequence=%d, want 123456", ev.LatestLedger, ev.LedgerSequence)
		}
	})

	t.Run("ledger_sequence only", func(t *testing.T) {
		raw := `{"source":"horizon","url":"https://h","claim":"c","ledger_sequence":123456}`
		var ev mechanics.Evidence
		if err := json.Unmarshal([]byte(raw), &ev); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if ev.LatestLedger != 123456 || ev.LedgerSequence != 123456 {
			t.Errorf("got LatestLedger=%d, LedgerSequence=%d, want 123456", ev.LatestLedger, ev.LedgerSequence)
		}
	})

	t.Run("neither", func(t *testing.T) {
		raw := `{"source":"stellar.expert","url":"https://se","claim":"c"}`
		var ev mechanics.Evidence
		if err := json.Unmarshal([]byte(raw), &ev); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if ev.LatestLedger != 0 || ev.LedgerSequence != 0 {
			t.Errorf("got LatestLedger=%d, LedgerSequence=%d, want 0", ev.LatestLedger, ev.LedgerSequence)
		}
		marshaled, err := json.Marshal(ev)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if strings.Contains(string(marshaled), "latest_ledger") || strings.Contains(string(marshaled), "ledger_sequence") {
			t.Errorf("zero ledger sequence should be omitted by omitempty: %s", string(marshaled))
		}
	})
}
