package cmd

import (
	"context"
	"errors"
	"time"

	"nmsmods/internal/app"
	"nmsmods/internal/nexus"

	"github.com/spf13/cobra"
)

var nexusGameDomain string

var nexusCmd = &cobra.Command{
	Use:   "nexus",
	Short: "Interact with the Nexus Mods API (info, files, auth, nxm)",
	Long:  "Interact with the Nexus Mods API (info, files, auth, nxm).",
}

func init() {
	// Default to the NMS Nexus game domain.
	nexusCmd.PersistentFlags().StringVar(&nexusGameDomain, "game", "nomanssky", "Nexus game domain (e.g. nomanssky)")
}

// nexusPathsConfig loads paths+config (no game requirement).
func nexusPathsConfig() (*app.Paths, app.Config, error) {
	p := mustPaths()
	cfg, err := loadConfig(p)
	if err != nil {
		return nil, app.Config{}, err
	}
	return p, cfg, nil
}

// requireNexusAPIKey returns config with a non-empty Nexus API key.
func requireNexusAPIKey(cfg app.Config) (string, error) {
	if cfg.Nexus.APIKey == "" {
		return "", errors.New("nexus api key not set (run: nmsmods nexus login)")
	}
	return cfg.Nexus.APIKey, nil
}

func newNexusClientFromConfig(cfg app.Config) (*nexus.Client, error) {
	key, err := requireNexusAPIKey(cfg)
	if err != nil {
		return nil, err
	}
	return nexus.NewClient(key, "nmsmods", app.Version), nil
}

func nexusCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
