package routing_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/routing"
)

func TestRule_Matches_Active(t *testing.T) {
	rule := &routing.Rule{IsActive: true, Currency: "USD", MinAmount: 100, MaxAmount: 10000}
	if !rule.Matches("USD", 5000) {
		t.Fatal("expected rule to match")
	}
}

func TestRule_Matches_Inactive(t *testing.T) {
	rule := &routing.Rule{IsActive: false}
	if rule.Matches("USD", 5000) {
		t.Fatal("inactive rule should not match")
	}
}

func TestRule_Matches_WrongCurrency(t *testing.T) {
	rule := &routing.Rule{IsActive: true, Currency: "EUR"}
	if rule.Matches("USD", 5000) {
		t.Fatal("should not match different currency")
	}
}

func TestRule_Matches_BelowMin(t *testing.T) {
	rule := &routing.Rule{IsActive: true, MinAmount: 1000}
	if rule.Matches("USD", 500) {
		t.Fatal("should not match below minimum amount")
	}
}

func TestRule_Matches_AboveMax(t *testing.T) {
	rule := &routing.Rule{IsActive: true, MaxAmount: 5000}
	if rule.Matches("USD", 10000) {
		t.Fatal("should not match above maximum amount")
	}
}

func TestRule_Matches_NoCurrencyFilter(t *testing.T) {
	rule := &routing.Rule{IsActive: true, Currency: ""}
	if !rule.Matches("USD", 5000) {
		t.Fatal("empty currency should match any")
	}
}

func TestRule_Matches_NoAmountFilter(t *testing.T) {
	rule := &routing.Rule{ID: uuid.New(), IsActive: true}
	if !rule.Matches("EUR", 999999) {
		t.Fatal("no amount filters should match any amount")
	}
}
