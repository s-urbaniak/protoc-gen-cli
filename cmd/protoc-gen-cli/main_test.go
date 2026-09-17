package main

import "testing"

func TestValidClient(t *testing.T) {
	t.Parallel()
	for _, client := range []string{"grpc-go", "connect-go"} {
		if !validClient(client) {
			t.Errorf("validClient(%q) = false, want true", client)
		}
	}
	for _, client := range []string{"", "grpc", "connect"} {
		if validClient(client) {
			t.Errorf("validClient(%q) = true, want false", client)
		}
	}
}
