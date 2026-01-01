package services

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaroslav/elias/internal/models"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	// This would normally connect to a test database
	// For now, we'll skip this and just test the logic
	t.Skip("Skipping database integration test - requires test database setup")
	return nil
}

func TestCreateRoomWithTeamNames(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	service := NewRoomService(pool)
	ctx := context.Background()

	user := &models.TelegramUser{
		ID:        12345,
		Username:  "testuser",
		FirstName: "Test",
	}

	tests := []struct {
		name     string
		category string
		numTeams int
		wantTeams int
	}{
		{"Create room with 2 teams", "general", 2, 2},
		{"Create room with 3 teams", "animals", 3, 3},
		{"Create room with 5 teams", "food", 5, 5},
		{"Create room with invalid teams (too many)", "general", 10, 5}, // Should clamp to 5
		{"Create room with invalid teams (too few)", "general", 1, 2},   // Should default to 2
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, player, err := service.CreateRoom(ctx, user, tt.category, tt.numTeams)
			if err != nil {
				t.Fatalf("CreateRoom failed: %v", err)
			}

			// Check room
			if room == nil {
				t.Fatal("Room is nil")
			}
			if room.Category != tt.category && tt.category != "" {
				t.Errorf("Expected category %s, got %s", tt.category, room.Category)
			}
			if room.NumTeams != tt.wantTeams {
				t.Errorf("Expected %d teams, got %d", tt.wantTeams, room.NumTeams)
			}
			if len(room.TeamNames) != tt.wantTeams {
				t.Errorf("Expected %d team names, got %d", tt.wantTeams, len(room.TeamNames))
			}

			// Check team names are unique
			seen := make(map[string]bool)
			for _, name := range room.TeamNames {
				if seen[name] {
					t.Errorf("Duplicate team name found: %s", name)
				}
				seen[name] = true
			}

			// Check player
			if player == nil {
				t.Fatal("Player is nil")
			}
			if !player.IsHost {
				t.Error("Creator should be host")
			}
			if player.UserID != user.ID {
				t.Errorf("Expected user ID %d, got %d", user.ID, player.UserID)
			}
		})
	}
}

func TestGetRoomWithTeamNames(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	service := NewRoomService(pool)
	ctx := context.Background()

	user := &models.TelegramUser{
		ID:        12345,
		Username:  "testuser",
		FirstName: "Test",
	}

	// Create a room
	room, _, err := service.CreateRoom(ctx, user, "general", 3)
	if err != nil {
		t.Fatalf("CreateRoom failed: %v", err)
	}

	// Retrieve the room
	retrievedRoom, err := service.GetRoom(ctx, room.ID)
	if err != nil {
		t.Fatalf("GetRoom failed: %v", err)
	}

	// Verify team names are retrieved correctly
	if len(retrievedRoom.TeamNames) != 3 {
		t.Errorf("Expected 3 team names, got %d", len(retrievedRoom.TeamNames))
	}

	// Verify team names match
	for i, name := range retrievedRoom.TeamNames {
		if name != room.TeamNames[i] {
			t.Errorf("Team name %d mismatch: expected %s, got %s", i, room.TeamNames[i], name)
		}
	}
}

func TestTeamNamesFormat(t *testing.T) {
	// Test that team names have the correct format (adjective + noun)
	tests := []struct {
		numTeams int
	}{
		{2},
		{3},
		{4},
		{5},
	}

	for _, tt := range tests {
		t.Run("Generate team names", func(t *testing.T) {
			names := GenerateUniqueTeamNames(tt.numTeams)

			for _, name := range names {
				// Each name should have exactly one space
				parts := len(name) - len(strings.Replace(name, " ", "", -1))
				if parts != 1 {
					t.Errorf("Expected team name to have 1 space, got %d: %s", parts, name)
				}

				// Name should not be empty
				if name == "" {
					t.Error("Team name is empty")
				}
			}
		})
	}
}
