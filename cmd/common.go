package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"nmsmods/internal/app"
	"nmsmods/internal/nms"
	"nmsmods/internal/steam"
)

func mustPaths() *app.Paths {
	p, err := app.DefaultPathsWithOverride(homeOverride)
	if err != nil {
		panic(err)
	}
	if err := p.Ensure(); err != nil {
		panic(err)
	}
	return p
}

func loadConfigAndMaybeGuess(p *app.Paths) (app.Config, error) {
	cfg, err := app.LoadConfig(p.Config)
	if err != nil {
		return cfg, err
	}
	if cfg.GamePath != "" {
		return cfg, nil
	}
	guesses, err := steam.GuessNMSPaths()
	if err != nil {
		return cfg, err
	}
	if len(guesses) > 0 {
		cfg.GamePath = guesses[0]
		_ = app.SaveConfig(p.Config, cfg) // best-effort
	}
	return cfg, nil
}

func requireGame(p *app.Paths) (*app.Config, *nms.Game, error) {
	cfg, err := loadConfigAndMaybeGuess(p)
	if err != nil {
		return nil, nil, err
	}
	if cfg.GamePath == "" {
		return &cfg, nil, errors.New("game path not set. Use: nmsmods set-path <path>")
	}
	game, err := nms.ValidateGamePath(cfg.GamePath)
	if err != nil {
		return &cfg, nil, fmt.Errorf("invalid game path: %w", err)
	}
	if err := nms.EnsureModsDir(game); err != nil {
		return &cfg, nil, err
	}
	return &cfg, game, nil
}

func isZipFile(name string) bool {
	l := strings.ToLower(name)
	return strings.HasSuffix(l, ".zip")
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// sortedModIDs returns stable ordering for numeric indexes.
func sortedModIDs(st app.State) []string {
	ids := make([]string, 0, len(st.Mods))
	for id := range st.Mods {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// resolveModArg allows commands to accept either an id (slug) or a numeric index from `downloads`.
func resolveModArg(arg string, st app.State) (string, error) {
	if n, err := strconv.Atoi(arg); err == nil {
		if n <= 0 {
			return "", fmt.Errorf("invalid index: %d", n)
		}
		ids := sortedModIDs(st)
		if n > len(ids) {
			return "", fmt.Errorf("index out of range: %d (have %d)", n, len(ids))
		}
		return ids[n-1], nil
	}
	if _, ok := st.Mods[arg]; ok {
		return arg, nil
	}
	return "", fmt.Errorf("unknown id: %s (run: nmsmods downloads)", arg)
}

func joinPathFromState(root, rel string) string {
	return filepath.Join(root, filepath.FromSlash(rel))
}

// withStateLock ensures we don't corrupt config/state when multiple nmsmods processes run.
// Wrap any command that writes config/state or deletes managed download assets.
func withStateLock(p *app.Paths, fn func() error) error {
	l, err := app.AcquireLock(p.Root)
	if err != nil {
		return err
	}
	defer l.Release()
	return fn()
}
