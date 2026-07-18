package provider_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paymentbridge/pcp/internal/domain/provider"
)

func TestProvider_Validate_Valid(t *testing.T) {
	p := &provider.Provider{ID: uuid.New(), Name: "Stripe Prod", Type: provider.TypeStripe}
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestProvider_Validate_EmptyName(t *testing.T) {
	p := &provider.Provider{ID: uuid.New(), Name: "", Type: provider.TypeStripe}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestProvider_Validate_InvalidType(t *testing.T) {
	p := &provider.Provider{ID: uuid.New(), Name: "Test", Type: "unknown"}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestProviderType_IsValid(t *testing.T) {
	cases := []struct {
		t     provider.Type
		valid bool
	}{
		{provider.TypeStripe, true},
		{provider.TypePayPal, true},
		{provider.TypeAdyen, true},
		{provider.TypeRazorpay, true},
		{"unknown", false},
		{"", false},
	}
	for _, c := range cases {
		if got := c.t.IsValid(); got != c.valid {
			t.Errorf("Type(%q).IsValid() = %v, want %v", c.t, got, c.valid)
		}
	}
}
