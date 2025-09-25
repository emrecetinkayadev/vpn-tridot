package middleware

import "testing"

func TestIPAllowedExactMatch(t *testing.T) {
	if !ipAllowed("203.0.113.10", []string{"203.0.113.10"}) {
		t.Fatalf("expected IP to be allowed")
	}
}

func TestIPAllowedCIDR(t *testing.T) {
	if !ipAllowed("10.0.0.5", []string{"10.0.0.0/24"}) {
		t.Fatalf("expected IP within CIDR to be allowed")
	}
}

func TestIPAllowedReject(t *testing.T) {
	if ipAllowed("192.0.2.50", []string{"10.0.0.0/24", "198.51.100.10"}) {
		t.Fatalf("expected IP to be rejected")
	}
}

func TestIPAllowedInvalidCIDR(t *testing.T) {
	if ipAllowed("192.0.2.50", []string{"invalid"}) {
		t.Fatalf("invalid CIDR should not allow access")
	}
}
