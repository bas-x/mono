package main

import (
	"testing"

	"github.com/bas-x/basex/services"
)

func TestBootstrapWorkspaceStartsWithBaseSimulationTab(t *testing.T) {
	t.Parallel()

	workspace, err := bootstrapWorkspace(services.NewSimulationService(services.SimulationServiceConfig{}))
	if err != nil {
		t.Fatalf("bootstrapWorkspace returned error: %v", err)
	}
	if workspace.activeTabID != services.BaseSimulationID {
		t.Fatalf("expected active tab %q, got %q", services.BaseSimulationID, workspace.activeTabID)
	}
	if workspace.baseInfo.ID != services.BaseSimulationID {
		t.Fatalf("expected base simulation info id %q, got %q", services.BaseSimulationID, workspace.baseInfo.ID)
	}
	if len(workspace.tabs) != 1 {
		t.Fatalf("expected exactly one startup tab, got %d", len(workspace.tabs))
	}
	if workspace.tabs[0].id != services.BaseSimulationID {
		t.Fatalf("expected startup tab id %q, got %q", services.BaseSimulationID, workspace.tabs[0].id)
	}
}

func TestBranchBaseTabAddsAndSelectsNewBranch(t *testing.T) {
	t.Parallel()

	service := services.NewSimulationService(services.SimulationServiceConfig{})
	if _, err := bootstrapWorkspace(service); err != nil {
		t.Fatalf("bootstrapWorkspace returned error: %v", err)
	}

	workspace, err := branchBaseTab(service)
	if err != nil {
		t.Fatalf("branchBaseTab returned error: %v", err)
	}
	if len(workspace.tabs) != 2 {
		t.Fatalf("expected two simulation tabs after branching, got %d", len(workspace.tabs))
	}
	if workspace.tabs[0].id != services.BaseSimulationID {
		t.Fatalf("expected first tab %q, got %q", services.BaseSimulationID, workspace.tabs[0].id)
	}
	if workspace.activeTabID == services.BaseSimulationID {
		t.Fatalf("expected a newly selected branch tab, got active %q", workspace.activeTabID)
	}
	if workspace.tabs[1].id != workspace.activeTabID {
		t.Fatalf("expected second tab %q to be active, got %q", workspace.tabs[1].id, workspace.activeTabID)
	}
}
