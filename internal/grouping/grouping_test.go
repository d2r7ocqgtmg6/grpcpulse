package grouping_test

import (
	"sort"
	"testing"

	"github.com/example/grpcpulse/internal/grouping"
)

func TestAdd_And_Members(t *testing.T) {
	r := grouping.New()
	r.Add("eu", "svc-a:443")
	r.Add("eu", "svc-b:443")

	members, err := r.Members("eu")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sort.Strings(members)
	if len(members) != 2 || members[0] != "svc-a:443" || members[1] != "svc-b:443" {
		t.Fatalf("unexpected members: %v", members)
	}
}

func TestMembers_UnknownGroup(t *testing.T) {
	r := grouping.New()
	_, err := r.Members("missing")
	if err != grouping.ErrGroupNotFound {
		t.Fatalf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestRemove_Target(t *testing.T) {
	r := grouping.New()
	r.Add("eu", "svc-a:443")
	r.Add("eu", "svc-b:443")

	if err := r.Remove("eu", "svc-a:443"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	members, _ := r.Members("eu")
	if len(members) != 1 || members[0] != "svc-b:443" {
		t.Fatalf("unexpected members after remove: %v", members)
	}
}

func TestRemove_LastTargetDeletesGroup(t *testing.T) {
	r := grouping.New()
	r.Add("eu", "svc-a:443")
	r.Remove("eu", "svc-a:443") //nolint:errcheck

	groups := r.Groups()
	if len(groups) != 0 {
		t.Fatalf("expected no groups, got %v", groups)
	}
}

func TestRemove_NotInGroup(t *testing.T) {
	r := grouping.New()
	r.Add("eu", "svc-a:443")
	err := r.Remove("eu", "svc-z:443")
	if err != grouping.ErrTargetNotInGroup {
		t.Fatalf("expected ErrTargetNotInGroup, got %v", err)
	}
}

func TestRemove_GroupNotFound(t *testing.T) {
	r := grouping.New()
	err := r.Remove("missing", "svc-a:443")
	if err != grouping.ErrGroupNotFound {
		t.Fatalf("expected ErrGroupNotFound, got %v", err)
	}
}

func TestGroups_ReturnsAllNames(t *testing.T) {
	r := grouping.New()
	r.Add("eu", "svc-a:443")
	r.Add("us", "svc-b:443")

	groups := r.Groups()
	sort.Strings(groups)
	if len(groups) != 2 || groups[0] != "eu" || groups[1] != "us" {
		t.Fatalf("unexpected groups: %v", groups)
	}
}
