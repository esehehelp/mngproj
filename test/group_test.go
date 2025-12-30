package test

import (
	"mngproj/pkg/config"
	"mngproj/pkg/manager"
	"testing"
)

func TestListComponentsByGroup(t *testing.T) {
	cfg := &config.ProjectConfig{
		Components: []config.ComponentConfig{
			{Name: "api", Groups: []string{"backend"}},
			{Name: "worker", Groups: []string{"backend", "async"}},
			{Name: "web", Groups: []string{"frontend"}},
		},
	}
	
	mgr := &manager.Manager{
		ProjectConfig: cfg,
	}
	
	// Test backend group
	backends := mgr.ListComponentsByGroup("backend")
	if len(backends) != 2 {
		t.Errorf("Expected 2 backend components, got %d", len(backends))
	}
	
	// Test async group
	asyncs := mgr.ListComponentsByGroup("async")
	if len(asyncs) != 1 {
		t.Errorf("Expected 1 async component, got %d", len(asyncs))
	}
	if len(asyncs) > 0 && asyncs[0] != "worker" {
		t.Errorf("Expected worker, got %s", asyncs[0])
	}
	
	// Test non-existent group
	none := mgr.ListComponentsByGroup("foo")
	if len(none) != 0 {
		t.Errorf("Expected 0 components, got %d", len(none))
	}
}
