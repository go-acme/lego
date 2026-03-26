package storage

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-acme/lego/v5/cmd/internal/configuration"
	"github.com/go-acme/lego/v5/log"
	"github.com/mattn/go-zglob"
)

func (m *Archiver) Accounts(cfg *configuration.Configuration) error {
	err := m.cleanArchivedAccounts()
	if err != nil {
		return fmt.Errorf("clean archived accounts: %w", err)
	}

	err = m.archiveAccounts(cfg)
	if err != nil {
		return fmt.Errorf("archive accounts: %w", err)
	}

	return nil
}

func (m *Archiver) ListArchivedAccounts() ([]string, error) {
	return listArchives(m.accountsArchivePath)
}

func (m *Archiver) archiveAccounts(cfg *configuration.Configuration) error {
	_, err := os.Stat(m.accountsBasePath)
	if os.IsNotExist(err) {
		return nil
	}

	matches, err := zglob.Glob(filepath.Join(m.accountsBasePath, "**", accountFileName))
	if err != nil {
		return fmt.Errorf("search account files: %w", err)
	}

	accountTree, err := accountMapping(cfg)
	if err != nil {
		return fmt.Errorf("account mapping: %w", err)
	}

	date := strconv.FormatInt(time.Now().Unix(), 10)

	for _, filename := range matches {
		dirAcc, _ := filepath.Split(filename)
		dirSrv, accID := filepath.Split(filepath.Dir(dirAcc))
		_, srv := filepath.Split(filepath.Dir(dirSrv))

		if _, ok := accountTree[srv]; !ok {
			err = m.archiveAccount("server", dirSrv, srv, date)
			if err != nil {
				return fmt.Errorf("archive account (server) %q: %w", srv, err)
			}

			continue
		}

		if _, ok := accountTree[srv][accID]; !ok {
			err = m.archiveAccount("accountID", dirAcc, srv, accID, date)
			if err != nil {
				return fmt.Errorf("archive account (accountID) %q: %w", accID, err)
			}

			continue
		}
	}

	return nil
}

func (m *Archiver) archiveAccount(scope, dir string, parts ...string) error {
	dest := filepath.Join(m.accountsArchivePath, strings.Join(parts, "_")+".zip")

	log.Info("Archive account",
		slog.String("scope", scope),
		slog.String("filepath", dir),
		slog.String("archives", dest),
	)

	err := CreateNonExistingFolder(filepath.Dir(dest))
	if err != nil {
		return fmt.Errorf("could not check/create the accounts archive folder %q: %w", filepath.Dir(dest), err)
	}

	rel, err := filepath.Rel(m.basePath, dir)
	if err != nil {
		return err
	}

	err = compressDirectory(dest, dir, rel)
	if err != nil {
		return fmt.Errorf("compress account files: %w", err)
	}

	return os.RemoveAll(dir)
}

func (m *Archiver) cleanArchivedAccounts() error {
	_, err := os.Stat(m.accountsArchivePath)
	if os.IsNotExist(err) {
		return nil
	}

	return m.cleanArchives(filepath.Join(m.accountsArchivePath, "**", "*.zip"))
}

func accountMapping(cfg *configuration.Configuration) (map[string]map[string]struct{}, error) {
	// Server -> AccountID
	accountTree := make(map[string]map[string]struct{})

	for accID := range cfg.Accounts {
		serverConfig := configuration.GetServerConfig(cfg, accID)

		s, err := url.Parse(serverConfig.URL)
		if err != nil {
			return nil, err
		}

		server := sanitizeHost(s)

		if _, ok := accountTree[server]; !ok {
			accountTree[server] = make(map[string]struct{})
		}

		accountTree[server][accID] = struct{}{}
	}

	return accountTree, nil
}
