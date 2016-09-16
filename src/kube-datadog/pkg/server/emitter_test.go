package server

import "testing"

func TestParseComputeResourceCPU(t *testing.T) {
	input := "100m"
	want := 100
	got, err := parseComputeResourceCPU(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want != got {
		t.Fatalf("incorrect value: want=%v; got=%v", want, got)
	}
}

func TestParseComputeResourceMemory(t *testing.T) {
	input := "64Mi"
	want := 67108864
	got, err := parseComputeResourceMemory(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want != got {
		t.Fatalf("incorrect value: want=%v; got=%v", want, got)
	}
}
