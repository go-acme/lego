package hook

import (
	"context"
	"fmt"
	"time"

	"github.com/go-acme/lego/v5/certcrypto"
	"github.com/go-acme/lego/v5/certificate"
	"github.com/go-acme/lego/v5/cmd/internal/storage"
	"github.com/go-acme/lego/v5/log"
)

// Action represents a hook action.
type Action struct {
	Cmd     string
	Timeout time.Duration
}

// Manager manages hooks.
type Manager struct {
	certsStorage *storage.CertificatesStorage

	metadata map[string]string

	pre    *Action
	deploy *Action
	post   *Action
}

// NewManager creates a new hook Manager.
func NewManager(certsStorage *storage.CertificatesStorage, options ...Option) *Manager {
	m := &Manager{
		certsStorage: certsStorage,
		metadata:     make(map[string]string),
	}

	for _, option := range options {
		option(m)
	}

	return m
}

// PreForDomains runs the pre-hook if defined.
func (h *Manager) PreForDomains(ctx context.Context, certID string, request certificate.ObtainRequest) error {
	if h.pre == nil || h.pre.Cmd == "" {
		return nil
	}

	keyType, err := request.EffectiveKeyType()
	if err != nil {
		return err
	}

	return h.preLaunch(ctx, certID, request.Domains, keyType)
}

// PreForCSR runs the pre-hook if defined.
func (h *Manager) PreForCSR(ctx context.Context, certID string, request certificate.ObtainForCSRRequest) error {
	if h.pre == nil || h.pre.Cmd == "" {
		return nil
	}

	keyType, err := request.EffectiveKeyType()
	if err != nil {
		return err
	}

	return h.preLaunch(ctx, certID, certcrypto.ExtractDomainsCSR(request.CSR), keyType)
}

// Deploy runs the deploy-hook if defined.
func (h *Manager) Deploy(ctx context.Context, certRes *certificate.Resource, options *storage.SaveOptions) error {
	if h.deploy == nil || h.deploy.Cmd == "" {
		return nil
	}

	addCertificateMetadata(h.metadata, certRes.ID, certRes.Domains, certRes.KeyType)
	addCertificatePathsMetadata(h.metadata, certRes, h.certsStorage, options)

	err := Launch(ctx, h.deploy.Cmd, h.deploy.Timeout, h.metadata)
	if err != nil {
		log.Error("Deploy hook.", log.ErrorAttr(err))

		return fmt.Errorf("deploy hook: %w", err)
	}

	return nil
}

// Post runs the post-hook if defined.
// This must be called inside a defer statement to ensure the hook is always run.
func (h *Manager) Post(ctx context.Context) error {
	if h.post == nil || h.post.Cmd == "" {
		return nil
	}

	err := Launch(ctx, h.post.Cmd, h.post.Timeout, h.metadata)
	if err != nil {
		log.Error("Post hook.", log.ErrorAttr(err))

		return fmt.Errorf("post hook: %w", err)
	}

	return nil
}

func (h *Manager) preLaunch(ctx context.Context, certID string, domains []string, keyType certcrypto.KeyType) error {
	addCertificateMetadata(h.metadata, certID, domains, keyType)

	err := Launch(ctx, h.pre.Cmd, h.pre.Timeout, h.metadata)
	if err != nil {
		log.Error("Pre hook.", log.ErrorAttr(err))

		return fmt.Errorf("pre hook: %w", err)
	}

	return nil
}
