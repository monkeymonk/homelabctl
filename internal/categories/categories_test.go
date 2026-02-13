package categories

import (
	"testing"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		catName  string
		wantErr  bool
		wantName string
	}{
		{
			name:     "valid category - infra",
			catName:  "infra",
			wantErr:  false,
			wantName: "infra",
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
			name:     "valid category - other",
			catName:  "other",
			wantErr:  false,
			wantName: "other",
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
	infra, _ := Get("infra")
	automation, _ := Get("automation")
	media, _ := Get("media")
	other, _ := Get("other")

	// Test that order is correct: infra < automation < media < other
	if infra.Order >= automation.Order {
		t.Errorf("infra.Order (%d) should be < automation.Order (%d)", infra.Order, automation.Order)
	}

	if automation.Order >= media.Order {
		t.Errorf("automation.Order (%d) should be < media.Order (%d)", automation.Order, media.Order)
	}

	if media.Order >= other.Order {
		t.Errorf("media.Order (%d) should be < other.Order (%d)", media.Order, other.Order)
	}
}

func TestValidCategoryName(t *testing.T) {
	tests := []struct {
		name     string
		catName  string
		wantValid bool
	}{
		{
			name:     "valid - infra",
			catName:  "infra",
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
			name:     "valid - other",
			catName:  "other",
			wantValid: true,
		},
		{
			name:     "invalid - random",
			catName:  "random",
			wantValid: false,
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
	tests := []struct {
		name           string
		catName        string
		expectedKeys   []string
	}{
		{
			name:         "infra has restart default",
			catName:      "infra",
			expectedKeys: []string{"restart"},
		},
		{
			name:         "media has environment defaults",
			catName:      "media",
			expectedKeys: []string{"environment"},
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
	all := AllCategories

	if len(all) != 4 {
		t.Errorf("Expected 4 categories, got %d", len(all))
	}

	// Check that all expected categories are present
	expectedNames := []string{"infra", "media", "automation", "other"}
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
	tests := []struct {
		name        string
		catName     string
		wantDisplay string
	}{
		{
			name:        "infra display name",
			catName:     "infra",
			wantDisplay: "Infrastructure",
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
			name:        "other display name",
			catName:     "other",
			wantDisplay: "Other",
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
	// Just verify that categories have colors defined
	all := AllCategories

	for _, cat := range all {
		if cat.Color == "" {
			t.Errorf("Category %s has no color defined", cat.Name)
		}
	}
}

func TestInfraDefaults(t *testing.T) {
	infra, err := Get("infra")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Infra should have restart policy
	if restart, ok := infra.Defaults["restart"]; !ok {
		t.Error("Infra category should have restart default")
	} else if restart != "unless-stopped" {
		t.Errorf("Infra restart = %v, want unless-stopped", restart)
	}
}

func TestMediaDefaults(t *testing.T) {
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
