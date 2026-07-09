package user

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) EnsureUser(username string) error {
	_, err := user.Lookup(username)
	if err == nil {
		return nil
	}
	var unknownUserError user.UnknownUserError
	if !errors.As(err, &unknownUserError) {
		return fmt.Errorf("lookup user %q: %w", username, err)
	}

	cmd := exec.Command("useradd", "-m", "-s", "/bin/bash", "-U", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("useradd %q: %w: %s", username, err, strings.TrimSpace(string(output)))
	}

	return nil
}

func (m *Manager) EnsureSudoer(username string) error {
	path := "/etc/sudoers.d/10-ec2-agent-user"
	content := fmt.Sprintf("%s ALL=(ALL) NOPASSWD:ALL\n", username)

	current, err := os.ReadFile(path)
	if err == nil && string(current) == content {
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read sudoers file for %q: %w", username, err)
	}

	if err := os.WriteFile(path, []byte(content), 0o440); err != nil {
		return fmt.Errorf("write sudoers file for %q: %w", username, err)
	}

	cmd := exec.Command("visudo", "-cf", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("validate sudoers file for %q: %w: %s", username, err, strings.TrimSpace(string(output)))
	}

	return nil
}

func (m *Manager) SetupPassword(username, password string) error {
	if password == "" {
		password = "!"
	}

	cmd := exec.Command("chpasswd", "-e")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s:%s\n", username, password))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("chpasswd password for %q: %w: %s", username, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func (m *Manager) SetupSSHKey(username, key string) (err error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}

	u, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("lookup user %q: %w", username, err)
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("parse uid for %q: %w", username, err)
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("parse gid for %q: %w", username, err)
	}

	sshDir := filepath.Join(u.HomeDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		return fmt.Errorf("create .ssh directory for %q: %w", username, err)
	}
	if err := os.Chown(sshDir, uid, gid); err != nil {
		return fmt.Errorf("chown .ssh directory for %q: %w", username, err)
	}

	authKeysPath := filepath.Join(sshDir, "authorized_keys")
	content, err := os.ReadFile(authKeysPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read authorized_keys for %q: %w", username, err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == key {
			return nil
		}
	}

	f, err := os.OpenFile(authKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open authorized_keys for %q: %w", username, err)
	}

	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close authorized_keys for %q: %w", username, closeErr)
		}
	}()

	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err := f.WriteString("\n"); err != nil {
			return fmt.Errorf("write authorized_keys for %q: %w", username, err)
		}
	}
	if _, err := f.WriteString(key + "\n"); err != nil {
		return fmt.Errorf("write authorized_keys for %q: %w", username, err)
	}

	if err := os.Chown(authKeysPath, uid, gid); err != nil {
		return fmt.Errorf("chown authorized_keys for %q: %w", username, err)
	}

	return nil
}
