package mods

import (
	"testing"

	"nmsmods/internal/app"
)

func TestResolveFolderCollision_ProfileScoped(t *testing.T) {
	st := app.State{Mods: map[string]app.ModEntry{
		"a": {Installations: map[string]app.ProfileInstall{"p1": {Installed: true, Folder: "Foo"}}},
		"b": {Installations: map[string]app.ProfileInstall{"p2": {Installed: true, Folder: "Foo"}}},
	}}

	folder, collided := ResolveFolderCollision("c", "Foo", "p1", st)
	if !collided {
		t.Fatalf("expected collision in p1")
	}
	if folder == "Foo" {
		t.Fatalf("expected renamed folder")
	}

	folder2, collided2 := ResolveFolderCollision("c", "Foo", "p3", st)
	if collided2 {
		t.Fatalf("did not expect collision in p3")
	}
	if folder2 != "Foo" {
		t.Fatalf("expected Foo, got %s", folder2)
	}
}
