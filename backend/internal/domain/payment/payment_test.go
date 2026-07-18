package payment_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/payment"
)

func TestPayment_Validate_Valid(t *testing.T) {
	p := &payment.Payment{ID: uuid.New(), MerchantID: uuid.New(), Amount: 5000, Currency: "USD"}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestPayment_Validate_ZeroAmount(t *testing.T) {
	p := &payment.Payment{ID: uuid.New(), MerchantID: uuid.New(), Amount: 0, Currency: "USD"}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestPayment_Validate_NegativeAmount(t *testing.T) {
	p := &payment.Payment{ID: uuid.New(), MerchantID: uuid.New(), Amount: -100, Currency: "USD"}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestPayment_Validate_InvalidCurrency(t *testing.T) {
	p := &payment.Payment{ID: uuid.New(), MerchantID: uuid.New(), Amount: 5000, Currency: "US"}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for 2-letter currency")
	}
}

func TestPayment_Validate_MissingMerchant(t *testing.T) {
	p := &payment.Payment{Amount: 5000, Currency: "USD"}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for missing merchant ID")
	}
}

func TestPayment_StatusTransitions(t *testing.T) {
	p := &payment.Payment{ID: uuid.New(), MerchantID: uuid.New(), Amount: 5000, Currency: "USD", Status: payment.StatusPending}

	provID := uuid.New()
	p.MarkProcessing(provID)
	if p.Status != payment.StatusProcessing {
		t.Fatalf("expected processing, got %s", p.Status)
	}
	if p.ProviderID != provID {
		t.Fatal("provider ID not set")
	}
	if p.AttemptCount != 1 {
		t.Fatal("attempt count not incremented")
	}

	p.MarkCompleted("ext_123")
	if p.Status != payment.StatusCompleted || p.ExternalID != "ext_123" {
		t.Fatal("mark completed failed")
	}

	p.MarkRefunded()
	if p.Status != payment.StatusRefunded {
		t.Fatal("mark refunded failed")
	}
}

func TestPayment_MarkFailed(t *testing.T) {
	p := &payment.Payment{Status: payment.StatusProcessing}
	p.MarkFailed("provider timeout")
	if p.Status != payment.StatusFailed || p.ErrorMessage != "provider timeout" {
		t.Fatal("mark failed did not set fields correctly")
	}
}
