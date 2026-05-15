package oncall

import (
	"testing"
	"time"
)

func baseRotation() Rotation {
	return Rotation{
		Name:          "platform",
		Responders:    []string{"alice", "bob", "carol"},
		ShiftDuration: time.Hour,
		StartedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestCurrentResponder_FirstShift(t *testing.T) {
	r := baseRotation()
	now := r.StartedAt.Add(30 * time.Minute)
	if got := r.CurrentResponder(now); got != "alice" {
		t.Fatalf("expected alice, got %s", got)
	}
}

func TestCurrentResponder_SecondShift(t *testing.T) {
	r := baseRotation()
	now := r.StartedAt.Add(90 * time.Minute)
	if got := r.CurrentResponder(now); got != "bob" {
		t.Fatalf("expected bob, got %s", got)
	}
}

func TestCurrentResponder_Wraps(t *testing.T) {
	r := baseRotation()
	// 3 responders × 1 h each → index 3 wraps to alice
	now := r.StartedAt.Add(3 * time.Hour)
	if got := r.CurrentResponder(now); got != "alice" {
		t.Fatalf("expected alice after wrap, got %s", got)
	}
}

func TestCurrentResponder_Empty(t *testing.T) {
	r := Rotation{Name: "empty", ShiftDuration: time.Hour, StartedAt: time.Now()}
	if got := r.CurrentResponder(time.Now()); got != "" {
		t.Fatalf("expected empty string for empty rotation, got %s", got)
	}
}

func TestSet_And_Get(t *testing.T) {
	reg := New()
	if err := reg.Set(baseRotation()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := reg.Get("platform")
	if !ok {
		t.Fatal("expected rotation to be found")
	}
	if got.Name != "platform" {
		t.Fatalf("unexpected name: %s", got.Name)
	}
}

func TestSet_ValidationErrors(t *testing.T) {
	reg := New()
	cases := []struct {
		name string
		r    Rotation
	}{
		{"empty name", Rotation{Responders: []string{"a"}, ShiftDuration: time.Hour}},
		{"no responders", Rotation{Name: "x", ShiftDuration: time.Hour}},
		{"zero duration", Rotation{Name: "x", Responders: []string{"a"}}},
	}
	for _, tc := range cases {
		if err := reg.Set(tc.r); err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
		}
	}
}

func TestDelete_RemovesRotation(t *testing.T) {
	reg := New()
	_ = reg.Set(baseRotation())
	reg.Delete("platform")
	if _, ok := reg.Get("platform"); ok {
		t.Fatal("expected rotation to be absent after delete")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	reg := New()
	_ = reg.Set(baseRotation())
	all := reg.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 rotation, got %d", len(all))
	}
	// mutating the returned slice must not affect the registry
	all[0].Name = "mutated"
	got, _ := reg.Get("platform")
	if got.Name != "platform" {
		t.Fatal("registry was mutated through returned slice")
	}
}
