package hook

import (
	"time"

	"github.com/go-acme/lego/v5/cmd/internal/storage"
)

type Option func(m *Manager)

// WithPre sets the pre-hook.
func WithPre(cmd string, timeout time.Duration) Option {
	return func(m *Manager) {
		if cmd == "" {
			return
		}

		m.pre = &Action{
			Cmd:     cmd,
			Timeout: timeout,
		}
	}
}

// WithDeploy sets the deploy-hook.
func WithDeploy(cmd string, timeout time.Duration) Option {
	return func(m *Manager) {
		if cmd == "" {
			return
		}

		m.deploy = &Action{
			Cmd:     cmd,
			Timeout: timeout,
		}
	}
}

// WithPost sets the post-hook.
func WithPost(cmd string, timeout time.Duration) Option {
	return func(m *Manager) {
		if cmd == "" {
			return
		}

		m.post = &Action{
			Cmd:     cmd,
			Timeout: timeout,
		}
	}
}

// WithAccountMetadata initializes the metadata with the account data.
func WithAccountMetadata(account *storage.Account) Option {
	return func(m *Manager) {
		if account == nil {
			return
		}

		addAccountMetadata(m.metadata, account)
	}
}
