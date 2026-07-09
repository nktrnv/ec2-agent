package main

import (
	"log/slog"
	"os"

	"github.com/C2Devel/ec2-agent/internal/imds"
	"github.com/C2Devel/ec2-agent/internal/user"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	meta, err := imds.NewClient().Fetch()
	if err != nil {
		logger.Error("failed to fetch metadata from IMDS", "error", err)
		os.Exit(1)
	}

	if meta.SSHKey == "" && meta.Password == "" {
		return
	}

	if meta.Username == "" {
		meta.Username = "ec2-user"
	}

	manager := user.NewManager()

	if err := manager.EnsureUser(meta.Username); err != nil {
		logger.Error("failed to ensure user", "username", meta.Username, "error", err)
		os.Exit(1)
	}

	if err := manager.EnsureSudoer(meta.Username); err != nil {
		logger.Error("failed to ensure sudoer", "username", meta.Username, "error", err)
		os.Exit(1)
	}

	if err := manager.SetupPassword(meta.Username, meta.Password); err != nil {
		logger.Error("failed to setup password", "username", meta.Username, "error", err)
		os.Exit(1)
	}

	if err := manager.SetupSSHKey(meta.Username, meta.SSHKey); err != nil {
		logger.Error("failed to setup ssh key", "username", meta.Username, "error", err)
		os.Exit(1)
	}

	logger.Info("login credentials setup completed successfully", "username", meta.Username)
}
