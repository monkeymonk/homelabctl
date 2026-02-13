package categories

import (
	"testing"
)

// setupTest registers test categories
func setupTest(t *testing.T) {
	t.Helper()
	Reset()
	RegisterCategory("core")
	RegisterCategory("infrastructure")
	RegisterCategory("monitoring")
	RegisterCategory("automation")
	RegisterCategory("media")
	RegisterCategory("tools")
}

func TestGet(t *testing.T) {
	setupTest(t)
	tests := []struct {
		name     string
		catName  string
		wantErr  bool
		wantName string
	}{
		{
			name:     "valid category - core",
			catName:  "core",
			wantErr:  false,
			wantName: "core",
		},
		{
			name:     "valid category - infrastructure",
			catName:  "infrastructure",
			wantErr:  false,
			wantName: "infrastructure",
		},
		{
			name:     "valid category - media",
			catName:  "media",
			wantErr:  false,
			wantName: "media",
		},
		{
			name:     "valid category - automation",
			catName:  "automation",
			wantErr:  false,
			wantName: "automation",
		},
		{
			name:     "valid category - tools",
			catName:  "tools",
			wantErr:  false,
			wantName: "tools",
		},
		{
			name:    "invalid category",
			catName: "nonexistent",
			wantErr: true,
		},
		{
			name:    "empty string",
			catName: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := Get(tt.catName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cat == nil {
					t.Fatal("Get() returned nil category")
				}
				if cat.Name != tt.wantName {
					t.Errorf("Get() name = %v, want %v", cat.Name, tt.wantName)
				}
			}
		})
	}
}

func TestCategoryOrder(t *testing.T) {
	setupTest(t)
	core, _ := Get("core")
	infrastructure, _ := Get("infrastructure")
	monitoring, _ := Get("monitoring")
	automation, _ := Get("automation")
	media, _ := Get("media")
	tools, _ := Get("tools")

	// Test that order is correct: core < infrastructure < monitoring < automation < media < tools
	if core.Order >= infrastructure.Order {
		t.Errorf("core.Order (%d) should be < infrastructure.Order (%d)", core.Order, infrastructure.Order)
	}

	if infrastructure.Order >= monitoring.Order {
		t.Errorf("infrastructure.Order (%d) should be < monitoring.Order (%d)", infrastructure.Order, monitoring.Order)
	}

	if monitoring.Order >= automation.Order {
		t.Errorf("monitoring.Order (%d) should be < automation.Order (%d)", monitoring.Order, automation.Order)
	}

	if automation.Order >= media.Order {
		t.Errorf("automation.Order (%d) should be < media.Order (%d)", automation.Order, media.Order)
	}

	if media.Order >= tools.Order {
		t.Errorf("media.Order (%d) should be < tools.Order (%d)", media.Order, tools.Order)
	}
}

func TestValidCategoryName(t *testing.T) {
	setupTest(t)
	tests := []struct {
		name     string
		catName  string
		wantValid bool
	}{
		{
			name:     "valid - core",
			catName:  "core",
			wantValid: true,
		},
		{
			name:     "valid - infrastructure",
			catName:  "infrastructure",
			wantValid: true,
		},
		{
			name:     "valid - media",
			catName:  "media",
			wantValid: true,
		},
		{
			name:     "valid - automation",
			catName:  "automation",
			wantValid: true,
		},
		{
			name:     "valid - random (dynamic discovery)",
			catName:  "random",
			wantValid: true,
		},
		{
			name:     "invalid - empty",
			catName:  "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidCategoryName(tt.catName)
			if got != tt.wantValid {
				t.Errorf("ValidCategoryName() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestCategoryDefaults(t *testing.T) {
	setupTest(t)
	tests := []struct {
		name           string
		catName        string
		expectedKeys   []string
	}{
		{
			name:         "core has restart and security defaults",
			catName:      "core",
			expectedKeys: []string{"restart", "security_opt"},
		},
		{
			name:         "infrastructure has restart and security defaults",
			catName:      "infrastructure",
			expectedKeys: []string{"restart", "security_opt"},
		},
		{
			name:         "media has restart and environment defaults",
			catName:      "media",
			expectedKeys: []string{"restart", "environment"},
		},
		{
			name:         "automation has restart default",
			catName:      "automation",
			expectedKeys: []string{"restart"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := Get(tt.catName)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			if cat.Defaults == nil {
				if len(tt.expectedKeys) > 0 {
					t.Error("Expected defaults but got nil")
				}
				return
			}

			for _, key := range tt.expectedKeys {
				if _, exists := cat.Defaults[key]; !exists {
					t.Errorf("Expected default key %q not found in %s category", key, tt.catName)
				}
			}
		})
	}
}

func TestAllCategories(t *testing.T) {
	setupTest(t)
	all := AllCategories()

	if len(all) != 6 {
		t.Errorf("Expected 6 categories, got %d", len(all))
	}

	// Check that all expected categories are present
	expectedNames := []string{"core", "infrastructure", "monitoring", "automation", "media", "tools"}
	foundNames := make(map[string]bool)

	for _, cat := range all {
		foundNames[cat.Name] = true
	}

	for _, expected := range expectedNames {
		if !foundNames[expected] {
			t.Errorf("Expected category %q not found in AllCategories()", expected)
		}
	}

	// Verify categories are sorted by order
	for i := 1; i < len(all); i++ {
		if all[i-1].Order >= all[i].Order {
			t.Errorf("Categories not sorted by order: %s (%d) should be before %s (%d)",
				all[i-1].Name, all[i-1].Order, all[i].Name, all[i].Order)
		}
	}
}

func TestCategoryDisplayNames(t *testing.T) {
	setupTest(t)
	tests := []struct {
		name        string
		catName     string
		wantDisplay string
	}{
		{
			name:        "core display name",
			catName:     "core",
			wantDisplay: "Core",
		},
		{
			name:        "infrastructure display name",
			catName:     "infrastructure",
			wantDisplay: "Infrastructure",
		},
		{
			name:        "monitoring display name",
			catName:     "monitoring",
			wantDisplay: "Monitoring",
		},
		{
			name:        "media display name",
			catName:     "media",
			wantDisplay: "Media",
		},
		{
			name:        "automation display name",
			catName:     "automation",
			wantDisplay: "Automation",
		},
		{
			name:        "tools display name",
			catName:     "tools",
			wantDisplay: "Tools",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := Get(tt.catName)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			if cat.DisplayName != tt.wantDisplay {
				t.Errorf("DisplayName = %v, want %v", cat.DisplayName, tt.wantDisplay)
			}
		})
	}
}

func TestCategoryColors(t *testing.T) {
	setupTest(t)
	// Just verify that categories have colors defined
	all := AllCategories()

	for _, cat := range all {
		if cat.Color == "" {
			t.Errorf("Category %s has no color defined", cat.Name)
		}
	}
}

func TestInfrastructureDefaults(t *testing.T) {
	setupTest(t)
	infrastructure, err := Get("infrastructure")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Infrastructure should have restart policy
	if restart, ok := infrastructure.Defaults["restart"]; !ok {
		t.Error("Infrastructure category should have restart default")
	} else if restart != "unless-stopped" {
		t.Errorf("Infrastructure restart = %v, want unless-stopped", restart)
	}

	// Infrastructure should have security_opt
	if _, ok := infrastructure.Defaults["security_opt"]; !ok {
		t.Error("Infrastructure category should have security_opt default")
	}
}

func TestMediaDefaults(t *testing.T) {
	setupTest(t)
	media, err := Get("media")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Media should have environment with PUID and PGID
	if env, ok := media.Defaults["environment"]; !ok {
		t.Error("Media category should have environment default")
	} else {
		if envMap, ok := env.(map[string]string); ok {
			if puid, ok := envMap["PUID"]; !ok {
				t.Error("Media environment should have PUID")
			} else if puid != "1000" {
				t.Errorf("PUID = %v, want 1000", puid)
			}

			if pgid, ok := envMap["PGID"]; !ok {
				t.Error("Media environment should have PGID")
			} else if pgid != "1000" {
				t.Errorf("PGID = %v, want 1000", pgid)
			}
		} else {
			t.Errorf("environment is not map[string]string, got %T", env)
		}
	}
}
