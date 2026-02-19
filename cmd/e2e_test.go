package cmd

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"nmsmods/internal/app"
)

func writeFile(t *testing.T, p string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func makeFakeNMS(t *testing.T, root string) {
	t.Helper()
	// Minimal structure to satisfy ValidateGamePath:
	// <root>/GAMEDATA/PCBANKS/*.pak, <root>/Binaries/
	writeFile(t, filepath.Join(root, "GAMEDATA", "PCBANKS", "dummy.pak"), []byte("pak"))
	if err := os.MkdirAll(filepath.Join(root, "Binaries"), 0o755); err != nil {
		t.Fatal(err)
	}
}

type zipBuild struct {
	w *zip.Writer
	t *testing.T
}

func zipWriter(t *testing.T, f *os.File) *zipBuild {
	return &zipBuild{w: zip.NewWriter(f), t: t}
}

func (z *zipBuild) Add(name string, data []byte) {
	wr, err := z.w.Create(name)
	if err != nil {
		z.t.Fatal(err)
	}
	_, _ = wr.Write(data)
}

func (z *zipBuild) Close() {
	if err := z.w.Close(); err != nil {
		z.t.Fatal(err)
	}
}

func TestE2E_LocalZip_Install_Verify_Uninstall(t *testing.T) {
	tmp := t.TempDir()

	// Isolated nmsmods home
	home := filepath.Join(tmp, "home")
	paths := app.PathsFromRoot(home)
	if err := paths.Ensure(); err != nil {
		t.Fatal(err)
	}

	// Fake game
	game := filepath.Join(tmp, "NMS")
	makeFakeNMS(t, game)

	// Set path in config
	cfg := app.Config{GamePath: game}
	if err := app.SaveConfig(paths.Config, cfg); err != nil {
		t.Fatal(err)
	}

	// Create a local ZIP to import: Outer/Inner/file.EXML (double folder case)
	localZip := filepath.Join(tmp, "mod.zip")
	{
		f, err := os.Create(localZip)
		if err != nil {
			t.Fatal(err)
		}
		zw := zipWriter(t, f)
		zw.Add("Outer/Inner/TEST.EXML", []byte("<xml/>"))
		zw.Close()
		_ = f.Close()
	}

	var out, errOut bytes.Buffer

	// download (import local zip)
	if err := ExecuteWithArgs([]string{"--home", home, "download", localZip, "--id", "testmod"}, &out, &errOut); err != nil {
		t.Fatalf("download failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()

	// install by index
	if err := ExecuteWithArgs([]string{"--home", home, "install", "1"}, &out, &errOut); err != nil {
		t.Fatalf("install failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()

	// verify should succeed
	if err := ExecuteWithArgs([]string{"--home", home, "verify", "1"}, &out, &errOut); err != nil {
		t.Fatalf("verify failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()

	// uninstall should succeed
	if err := ExecuteWithArgs([]string{"--home", home, "uninstall", "1"}, &out, &errOut); err != nil {
		t.Fatalf("uninstall failed: %v\nout=%s\nerr=%s", err, out.String(), errOut.String())
	}
}
