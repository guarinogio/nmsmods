package mods

import (
	"testing"

	"nmsmods/internal/app"
)

func TestResolveFolderCollision(t *testing.T) {
	st := app.State{
		Mods: map[string]app.ModEntry{
			"other": {Installed: true, Folder: "Inner"},
		},
	}

	got, collided := ResolveFolderCollision("newmod", "Inner", st)
	if !collided {
		t.Fatalf("expected collision=true")
	}
	if got == "Inner" {
		t.Fatalf("expected different folder, got %q", got)
	}
}
