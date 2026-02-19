package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"nmsmods/internal/app"
)

func TestDoctor_FailsWhenGameInvalid(t *testing.T) {
	tmp := t.TempDir()

	home := filepath.Join(tmp, "home")
	paths := app.PathsFromRoot(home)
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}

	// invalid game path (does not exist)
	if err := app.SaveConfig(paths.Config, app.Config{GamePath: filepath.Join(tmp, "NOPE")}); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err := ExecuteWithArgs([]string{"--home", home, "doctor", "--json"}, &out, &errOut)
	if err == nil {
		t.Fatalf("expected doctor error, got nil\nout=%s\nerr=%s", out.String(), errOut.String())
	}
}

func TestDoctor_OkWhenGameValid(t *testing.T) {
	tmp := t.TempDir()

	home := filepath.Join(tmp, "home")
	paths := app.PathsFromRoot(home)
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}

	game := filepath.Join(tmp, "NMS")
	// minimal valid structure
	if err := os.MkdirAll(filepath.Join(game, "GAMEDATA", "PCBANKS"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(game, "Binaries"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(game, "GAMEDATA", "PCBANKS", "dummy.pak"), []byte("pak"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := app.SaveConfig(paths.Config, app.Config{GamePath: game}); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	if err := ExecuteWithArgs([]string{"--home", home, "doctor", "--json"}, &out, &errOut); err != nil {
		t.Fatalf("expected doctor ok, got err=%v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
}
