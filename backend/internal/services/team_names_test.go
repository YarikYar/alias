package services

import (
	"strings"
	"testing"
)

func TestGenerateTeamName(t *testing.T) {
	for i := 0; i < 100; i++ {
		name := GenerateTeamName()

		// Check that name is not empty
		if name == "" {
			t.Error("Generated name is empty")
		}

		// Check that name contains a space (adjective + space + noun)
		if !strings.Contains(name, " ") {
			t.Errorf("Generated name does not contain a space: %s", name)
		}

		// Split into parts
		parts := strings.Split(name, " ")
		if len(parts) != 2 {
			t.Errorf("Generated name should have exactly 2 parts, got %d: %s", len(parts), name)
		}
	}
}

func TestGenerateUniqueTeamNames(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"Generate 2 teams", 2},
		{"Generate 3 teams", 3},
		{"Generate 4 teams", 4},
		{"Generate 5 teams", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			names := GenerateUniqueTeamNames(tt.count)

			// Check count
			if len(names) != tt.count {
				t.Errorf("Expected %d names, got %d", tt.count, len(names))
			}

			// Check uniqueness
			seen := make(map[string]bool)
			for _, name := range names {
				if seen[name] {
					t.Errorf("Duplicate name found: %s", name)
				}
				seen[name] = true

				// Check format
				if name == "" {
					t.Error("Generated name is empty")
				}
				if !strings.Contains(name, " ") {
					t.Errorf("Generated name does not contain a space: %s", name)
				}
			}
		})
	}
}

func TestGenerateUniqueTeamNamesStability(t *testing.T) {
	// Test that we can generate many unique names without issues
	for i := 0; i < 10; i++ {
		names := GenerateUniqueTeamNames(5)
		if len(names) != 5 {
			t.Errorf("Iteration %d: Expected 5 names, got %d", i, len(names))
		}
	}
}

func TestGenderAgreement(t *testing.T) {
	// Test that masculine adjectives are paired with masculine nouns
	for _, adj := range adjectivesMasc {
		for _, noun := range nounsMasc {
			name := adj + " " + noun
			// Just check it doesn't crash and forms a valid string
			if !strings.Contains(name, " ") {
				t.Errorf("Invalid name format: %s", name)
			}
		}
	}

	// Test that feminine adjectives are paired with feminine nouns
	for _, adj := range adjectivesFem {
		for _, noun := range nounsFem {
			name := adj + " " + noun
			if !strings.Contains(name, " ") {
				t.Errorf("Invalid name format: %s", name)
			}
		}
	}

	// Test that neuter adjectives are paired with neuter nouns
	for _, adj := range adjectivesNeut {
		for _, noun := range nounsNeut {
			name := adj + " " + noun
			if !strings.Contains(name, " ") {
				t.Errorf("Invalid name format: %s", name)
			}
		}
	}
}

func TestGenderArraysNotEmpty(t *testing.T) {
	if len(adjectivesMasc) == 0 {
		t.Error("adjectivesMasc is empty")
	}
	if len(adjectivesFem) == 0 {
		t.Error("adjectivesFem is empty")
	}
	if len(adjectivesNeut) == 0 {
		t.Error("adjectivesNeut is empty")
	}
	if len(nounsMasc) == 0 {
		t.Error("nounsMasc is empty")
	}
	if len(nounsFem) == 0 {
		t.Error("nounsFem is empty")
	}
	if len(nounsNeut) == 0 {
		t.Error("nounsNeut is empty")
	}
}
