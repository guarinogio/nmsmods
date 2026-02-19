package cmd

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"nmsmods/internal/app"
)

func writeFileE2E(t *testing.T, p string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func makeFakeNMSE2E(t *testing.T, root string) {
	t.Helper()
	// Minimal structure to satisfy ValidateGamePath:
	// <root>/GAMEDATA/PCBANKS/*.pak, <root>/Binaries/
	writeFileE2E(t, filepath.Join(root, "GAMEDATA", "PCBANKS", "dummy.pak"), []byte("pak"))
	if err := os.MkdirAll(filepath.Join(root, "Binaries"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func mkZipE2E(t *testing.T, zipPath string, entries map[string][]byte) {
	t.Helper()
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, data := range entries {
		zf, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = zf.Write(data)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestE2E_OverwriteDefault(t *testing.T) {
	tmp := t.TempDir()

	home := filepath.Join(tmp, "home")
	paths := app.PathsFromRoot(home)
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}

	game := filepath.Join(tmp, "NMS")
	makeFakeNMSE2E(t, game)
	if err := app.SaveConfig(paths.Config, app.Config{GamePath: game}); err != nil {
		t.Fatal(err)
	}

	zip1 := filepath.Join(tmp, "mod_v1.zip")
	mkZipE2E(t, zip1, map[string][]byte{
		"Outer/Inner/TEST.EXML": []byte("v1"),
	})
	zip2 := filepath.Join(tmp, "mod_v2.zip")
	mkZipE2E(t, zip2, map[string][]byte{
		"Outer/Inner/TEST.EXML": []byte("v2"),
	})

	var out, errOut bytes.Buffer

	// import v1 under same id
	if err := ExecuteWithArgs([]string{"--home", home, "download", zip1, "--id", "testmod"}, &out, &errOut); err != nil {
		t.Fatalf("download v1 failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()

	// install
	if err := ExecuteWithArgs([]string{"--home", home, "install", "1"}, &out, &errOut); err != nil {
		t.Fatalf("install v1 failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()

	installedFile := filepath.Join(game, "GAMEDATA", "MODS", "Inner", "TEST.EXML")
	b1, err := os.ReadFile(installedFile)
	if err != nil {
		t.Fatalf("expected installed file: %v", err)
	}
	if string(b1) != "v1" {
		t.Fatalf("expected v1, got %q", string(b1))
	}

	// import v2 under same id (updates state.ZIP)
	if err := ExecuteWithArgs([]string{"--home", home, "download", zip2, "--id", "testmod"}, &out, &errOut); err != nil {
		t.Fatalf("download v2 failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()

	// install again (overwrite default)
	if err := ExecuteWithArgs([]string{"--home", home, "install", "1"}, &out, &errOut); err != nil {
		t.Fatalf("install v2 failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()

	b2, err := os.ReadFile(installedFile)
	if err != nil {
		t.Fatalf("expected installed file after overwrite: %v", err)
	}
	if string(b2) != "v2" {
		t.Fatalf("expected v2 after overwrite, got %q", string(b2))
	}
}

func TestE2E_NoOverwriteFlag(t *testing.T) {
	tmp := t.TempDir()

	home := filepath.Join(tmp, "home")
	paths := app.PathsFromRoot(home)
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}

	game := filepath.Join(tmp, "NMS")
	makeFakeNMSE2E(t, game)
	if err := app.SaveConfig(paths.Config, app.Config{GamePath: game}); err != nil {
		t.Fatal(err)
	}

	zip1 := filepath.Join(tmp, "mod.zip")
	mkZipE2E(t, zip1, map[string][]byte{
		"Outer/Inner/TEST.EXML": []byte("v1"),
	})

	var out, errOut bytes.Buffer

	if err := ExecuteWithArgs([]string{"--home", home, "download", zip1, "--id", "testmod"}, &out, &errOut); err != nil {
		t.Fatalf("download failed: %v", err)
	}
	out.Reset()
	errOut.Reset()

	if err := ExecuteWithArgs([]string{"--home", home, "install", "1"}, &out, &errOut); err != nil {
		t.Fatalf("install failed: %v", err)
	}
	out.Reset()
	errOut.Reset()

	// should fail because destination exists
	if err := ExecuteWithArgs([]string{"--home", home, "install", "1", "--no-overwrite"}, &out, &errOut); err == nil {
		t.Fatalf("expected error with --no-overwrite, got nil\nout=%s\nerr=%s", out.String(), errOut.String())
	}
}

func TestE2E_InstallDir(t *testing.T) {
	tmp := t.TempDir()

	home := filepath.Join(tmp, "home")
	paths := app.PathsFromRoot(home)
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}

	game := filepath.Join(tmp, "NMS")
	makeFakeNMSE2E(t, game)
	if err := app.SaveConfig(paths.Config, app.Config{GamePath: game}); err != nil {
		t.Fatal(err)
	}

	// local extracted dir: Outer/Inner/TEST.EXML
	src := filepath.Join(tmp, "srcmod")
	writeFileE2E(t, filepath.Join(src, "Outer", "Inner", "TEST.EXML"), []byte("dir"))

	var out, errOut bytes.Buffer

	if err := ExecuteWithArgs([]string{"--home", home, "install-dir", src, "--id", "dirmod"}, &out, &errOut); err != nil {
		t.Fatalf("install-dir failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}

	installedFile := filepath.Join(game, "GAMEDATA", "MODS", "Inner", "TEST.EXML")
	b, err := os.ReadFile(installedFile)
	if err != nil {
		t.Fatalf("expected installed file: %v", err)
	}
	if string(b) != "dir" {
		t.Fatalf("expected dir, got %q", string(b))
	}
}

func TestE2E_CleanAndReset(t *testing.T) {
	tmp := t.TempDir()

	home := filepath.Join(tmp, "home")
	paths := app.PathsFromRoot(home)
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}

	game := filepath.Join(tmp, "NMS")
	makeFakeNMSE2E(t, game)
	if err := app.SaveConfig(paths.Config, app.Config{GamePath: game}); err != nil {
		t.Fatal(err)
	}

	// Create a referenced zip via download (so state points to it)
	zip1 := filepath.Join(tmp, "mod.zip")
	mkZipE2E(t, zip1, map[string][]byte{"A/TEST.EXML": []byte("x")})

	var out, errOut bytes.Buffer
	if err := ExecuteWithArgs([]string{"--home", home, "download", zip1, "--id", "ref"}, &out, &errOut); err != nil {
		t.Fatalf("download failed: %v", err)
	}
	out.Reset()
	errOut.Reset()

	// Add an orphan zip directly in downloads/
	orphan := filepath.Join(paths.Downloads, "orphan.zip")
	mkZipE2E(t, orphan, map[string][]byte{"X/NO.EXML": []byte("y")})

	// Add a .part file
	part := filepath.Join(paths.Downloads, "file.zip.part")
	if err := os.WriteFile(part, []byte("partial"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Add something in staging
	staged := filepath.Join(paths.Staging, "tmp.txt")
	if err := os.WriteFile(staged, []byte("staging"), 0o644); err != nil {
		t.Fatal(err)
	}

	// clean parts + orphan + staging
	if err := ExecuteWithArgs([]string{"--home", home, "clean", "--staging", "--parts", "--orphan-zips"}, &out, &errOut); err != nil {
		t.Fatalf("clean failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()

	if _, err := os.Stat(part); err == nil {
		t.Fatalf("expected part file removed")
	}
	if _, err := os.Stat(orphan); err == nil {
		t.Fatalf("expected orphan zip removed")
	}
	// staging should exist (recreated) but the staged file should be gone
	if _, err := os.Stat(staged); err == nil {
		t.Fatalf("expected staged file removed")
	}

	// reset should remove state.json but keep downloads by default
	if err := ExecuteWithArgs([]string{"--home", home, "reset"}, &out, &errOut); err != nil {
		t.Fatalf("reset failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()

	if _, err := os.Stat(paths.State); err == nil {
		t.Fatalf("expected state.json removed by reset")
	}
	if _, err := os.Stat(paths.Downloads); err != nil {
		t.Fatalf("expected downloads kept by default: %v", err)
	}
}
