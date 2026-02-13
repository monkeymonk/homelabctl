package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeComposeFiles_Basic(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create first compose file
	file1 := filepath.Join(tmpDir, "stack1.yml")
	content1 := `services:
  app1:
    image: nginx:1
    container_name: app1
    restart: unless-stopped
volumes:
  app1_data: {}
networks:
  default: {}
`
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create second compose file
	file2 := filepath.Join(tmpDir, "stack2.yml")
	content2 := `services:
  app2:
    image: postgres:14
    container_name: app2
    restart: unless-stopped
volumes:
  app2_data: {}
networks:
  default: {}
`
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Merge
	merged, err := MergeComposeFiles([]string{file1, file2})
	if err != nil {
		t.Fatalf("MergeComposeFiles() unexpected error: %v", err)
	}

	// Check services from both files are present
	if len(merged.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(merged.Services))
	}

	if _, exists := merged.Services["app1"]; !exists {
		t.Error("Service app1 should exist in merged result")
	}

	if _, exists := merged.Services["app2"]; !exists {
		t.Error("Service app2 should exist in merged result")
	}

	// Check volumes are present
	if len(merged.Volumes) != 2 {
		t.Errorf("Expected 2 volumes, got %d", len(merged.Volumes))
	}

	// Check networks are deduplicated
	if len(merged.Networks) != 1 {
		t.Errorf("Expected 1 network (deduplicated), got %d", len(merged.Networks))
	}
}

func TestMergeComposeFiles_DuplicateService(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "stack1.yml")
	content1 := `services:
  app:
    image: nginx:1
`
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	file2 := filepath.Join(tmpDir, "stack2.yml")
	content2 := `services:
  app:
    image: nginx:2
`
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := MergeComposeFiles([]string{file1, file2})
	if err == nil {
		t.Fatal("Expected error for duplicate service name, got nil")
	}

	if !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "app") {
		t.Errorf("Error should mention duplicate service, got: %v", err)
	}
}

func TestMergeComposeFiles_VolumeDeduplication(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "stack1.yml")
	content1 := `services:
  app1:
    image: nginx:1
volumes:
  shared_data: {}
  app1_data: {}
`
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	file2 := filepath.Join(tmpDir, "stack2.yml")
	content2 := `services:
  app2:
    image: nginx:2
volumes:
  shared_data: {}
  app2_data: {}
`
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	merged, err := MergeComposeFiles([]string{file1, file2})
	if err != nil {
		t.Fatalf("MergeComposeFiles() unexpected error: %v", err)
	}

	// Should have 3 unique volumes
	if len(merged.Volumes) != 3 {
		t.Errorf("Expected 3 volumes (deduplicated), got %d", len(merged.Volumes))
	}

	expectedVolumes := []string{"shared_data", "app1_data", "app2_data"}
	for _, vol := range expectedVolumes {
		if _, exists := merged.Volumes[vol]; !exists {
			t.Errorf("Volume %s should exist in merged result", vol)
		}
	}
}

func TestMergeComposeFiles_NetworkDeduplication(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "stack1.yml")
	content1 := `services:
  app1:
    image: nginx:1
networks:
  default: {}
  traefik:
    external: true
`
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	file2 := filepath.Join(tmpDir, "stack2.yml")
	content2 := `services:
  app2:
    image: nginx:2
networks:
  default: {}
  databases:
    external: true
`
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	merged, err := MergeComposeFiles([]string{file1, file2})
	if err != nil {
		t.Fatalf("MergeComposeFiles() unexpected error: %v", err)
	}

	// Should have 3 unique networks
	if len(merged.Networks) != 3 {
		t.Errorf("Expected 3 networks (deduplicated), got %d", len(merged.Networks))
	}

	expectedNetworks := []string{"default", "traefik", "databases"}
	for _, net := range expectedNetworks {
		if _, exists := merged.Networks[net]; !exists {
			t.Errorf("Network %s should exist in merged result", net)
		}
	}
}

func TestMergeComposeFiles_Empty(t *testing.T) {
	tests := []struct {
		name   string
		files  []string
		wantOk bool
	}{
		{
			name:   "nil input",
			files:  nil,
			wantOk: true,
		},
		{
			name:   "empty slice",
			files:  []string{},
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged, err := MergeComposeFiles(tt.files)
			if tt.wantOk && err != nil {
				t.Errorf("MergeComposeFiles() unexpected error: %v", err)
			}
			if tt.wantOk && merged == nil {
				t.Error("MergeComposeFiles() should return non-nil result")
			}
		})
	}
}

func TestWriteComposeFile(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "docker-compose.yml")

	compose := &ComposeFile{
		Services: map[string]interface{}{
			"test": map[string]interface{}{
				"image":          "nginx:latest",
				"container_name": "test-container",
			},
		},
	}

	err := WriteComposeFile(outputPath, compose)
	if err != nil {
		t.Fatalf("WriteComposeFile() unexpected error: %v", err)
	}

	// Verify file was created and has content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "test") {
		t.Error("Output file should contain service name 'test'")
	}

	if !strings.Contains(content, "nginx:latest") {
		t.Error("Output file should contain image 'nginx:latest'")
	}
}

func TestFilterDisabledServices(t *testing.T) {
	tests := []struct {
		name            string
		disabled        []string
		wantRemoved     []string
		wantServicesLen int
	}{
		{
			name:            "no disabled services",
			disabled:        []string{},
			wantRemoved:     nil,
			wantServicesLen: 3,
		},
		{
			name:            "one disabled service",
			disabled:        []string{"disabled"},
			wantRemoved:     []string{"disabled"},
			wantServicesLen: 2,
		},
		{
			name:            "multiple disabled services",
			disabled:        []string{"app1", "disabled"},
			wantRemoved:     []string{"app1", "disabled"},
			wantServicesLen: 1,
		},
		{
			name:            "disable non-existent service",
			disabled:        []string{"nonexistent"},
			wantRemoved:     nil,
			wantServicesLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh copy for each test
			testCompose := &ComposeFile{
				Services: map[string]interface{}{
					"app1":     map[string]interface{}{"image": "nginx:1"},
					"app2":     map[string]interface{}{"image": "nginx:2"},
					"disabled": map[string]interface{}{"image": "nginx:3"},
				},
			}

			removed := FilterDisabledServices(testCompose, tt.disabled)

			if len(testCompose.Services) != tt.wantServicesLen {
				t.Errorf("Services length = %d, want %d", len(testCompose.Services), tt.wantServicesLen)
			}

			if len(removed) != len(tt.wantRemoved) {
				t.Errorf("Removed length = %d, want %d", len(removed), len(tt.wantRemoved))
			}

			// Check that disabled services were actually removed
			for _, svc := range tt.disabled {
				if _, exists := testCompose.Services[svc]; exists {
					if contains(tt.wantRemoved, svc) {
						t.Errorf("Service %s should have been removed", svc)
					}
				}
			}
		})
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
