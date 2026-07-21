package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStructuralIntegrityAndPersistence(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_budget_isolated.json")
	templateFile := filepath.Join(testDir, "index.html")
	if err := os.WriteFile(templateFile, []byte("<html></html>"), 0o644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	app := NewBudgetApp(testFile, templateFile)
	app.users["demo"] = User{Username: "demo", Currency: "KSh", Income: 250000}
	if err := app.saveUsers(); err != nil {
		t.Fatalf("failed to save users: %v", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected persisted content")
	}
}

func TestGreetingAndCurrencyHelpers(t *testing.T) {
	if got := greetingForCurrency("USD"); got != "Hello" {
		t.Fatalf("expected US greeting to be Hello, got %q", got)
	}
	if got := greetingForCurrency("JPY"); got != "こんにちは" {
		t.Fatalf("expected Japanese greeting to be こんにちは, got %q", got)
	}
	if got := greetingForCurrency("TZS"); got != "Karibu" {
		t.Fatalf("expected TZS greeting to be Karibu, got %q", got)
	}
	if got := greetingForCurrency("UGX"); got != "Oli otya" {
		t.Fatalf("expected UGX greeting to be Oli otya, got %q", got)
	}
	if symbol, label := currencyMeta("JPY"); symbol != "¥" || label == "" {
		t.Fatalf("expected JPY symbol and label, got %q %q", symbol, label)
	}
}

func TestCurrencyAdviceMatchesCurrencyContext(t *testing.T) {
	advice := currencyAdviceForCurrency("KES")
	if advice.SaccoTitle == "" || advice.MmfTitle == "" || advice.TBillTitle == "" {
		t.Fatalf("expected advice blocks for KES currency, got %+v", advice)
	}
	if advice.SaccoBody == "" || advice.MmfBody == "" || advice.TBillBody == "" {
		t.Fatalf("expected populated bodies for KES currency, got %+v", advice)
	}
	if !strings.Contains(advice.SaccoBody, "M-Pesa") && !strings.Contains(advice.MmfBody, "M-Pesa") {
		t.Fatalf("expected East African advice to mention M-Pesa or bank withdrawals, got %q", advice.SaccoBody)
	}

	globalAdvice := currencyAdviceForCurrency("XYZ")
	if globalAdvice.SaccoTitle == "" || globalAdvice.Intro == "" {
		t.Fatalf("expected fallback advice for unknown currency, got %+v", globalAdvice)
	}
}
