package merchant_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/merchant"
)

func TestMerchant_Validate_ValidMerchant(t *testing.T) {
	m := &merchant.Merchant{
		ID:     uuid.New(),
		Name:   "Test Merchant",
		APIKey: "pcp_test123",
		Status: merchant.StatusActive,
	}

	if err := m.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMerchant_Validate_EmptyName(t *testing.T) {
	m := &merchant.Merchant{
		ID:     uuid.New(),
		Name:   "",
		APIKey: "pcp_test123",
		Status: merchant.StatusActive,
	}

	err := m.Validate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if err != merchant.ErrInvalidName {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestMerchant_Validate_NameTooLong(t *testing.T) {
	longName := make([]byte, 256)
	for i := range longName {
		longName[i] = 'a'
	}
	m := &merchant.Merchant{
		ID:     uuid.New(),
		Name:   string(longName),
		APIKey: "pcp_test123",
		Status: merchant.StatusActive,
	}

	err := m.Validate()
	if err == nil {
		t.Fatal("expected error for long name")
	}
	if err != merchant.ErrInvalidName {
		t.Errorf("expected ErrInvalidName, got %v", err)
	}
}

func TestMerchant_Validate_InvalidStatus(t *testing.T) {
	m := &merchant.Merchant{
		ID:     uuid.New(),
		Name:   "Test",
		APIKey: "pcp_test123",
		Status: merchant.Status("invalid"),
	}

	err := m.Validate()
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if err != merchant.ErrInvalidStatus {
		t.Errorf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		status merchant.Status
		valid  bool
	}{
		{merchant.StatusActive, true},
		{merchant.StatusSuspended, true},
		{merchant.StatusInactive, true},
		{merchant.Status("unknown"), false},
		{merchant.Status(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.valid {
				t.Errorf("Status(%q).IsValid() = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestMerchant_Activate(t *testing.T) {
	m := &merchant.Merchant{Status: merchant.StatusSuspended}
	m.Activate()

	if m.Status != merchant.StatusActive {
		t.Errorf("expected StatusActive, got %v", m.Status)
	}
	if m.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestMerchant_Suspend(t *testing.T) {
	m := &merchant.Merchant{Status: merchant.StatusActive}
	m.Suspend()

	if m.Status != merchant.StatusSuspended {
		t.Errorf("expected StatusSuspended, got %v", m.Status)
	}
}

func TestMerchant_Deactivate(t *testing.T) {
	m := &merchant.Merchant{Status: merchant.StatusActive}
	m.Deactivate()

	if m.Status != merchant.StatusInactive {
		t.Errorf("expected StatusInactive, got %v", m.Status)
	}
}
