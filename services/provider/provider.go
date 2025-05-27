package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/model"
	db_provider "github.com/Yiling-J/tablepilot/ent/provider"
	"go.uber.org/zap"
)

//go:generate moq -rm -out provider_moq.go . ProviderService
type ProviderService interface {
	ListProviders(ctx context.Context) ([]Provider, error)
	CreateProvider(ctx context.Context, provider Provider) error
	UpdateProvider(ctx context.Context, id int, provider Provider) error
	DeleteProvider(ctx context.Context, id int) error
	WithSyncCallback(callback func(ctx context.Context, providers []Provider))
	BuildProviders(ctx context.Context) error
}

type ProviderServiceImpl struct {
	db           *ent.Client
	cfg          *config.Config
	logger       *zap.SugaredLogger
	mu           sync.Mutex
	providers    []Provider
	syncCallback func(ctx context.Context, providers []Provider)
}

func NewProviderService(cfg *config.Config, db *ent.Client, logger *zap.SugaredLogger) *ProviderServiceImpl {
	return &ProviderServiceImpl{
		db:     db,
		cfg:    cfg,
		logger: logger.With("service", "provider"),
	}
}

func (p *ProviderServiceImpl) CreateProvider(ctx context.Context, provider Provider) error {
	p.mu.Lock()
	defer func() {
		ps, err := p.genProviders(ctx)
		if err != nil {
			p.logger.Errorw("delete provider failed", "error", err)
		} else {
			p.providers = ps
		}
		if p.syncCallback != nil {
			p.syncCallback(ctx, ps)
		}
		p.mu.Unlock()
	}()
	tx, err := p.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("provider.CreateProvider: starting transaction: %w", err)
	}
	pv, err := tx.Provider.Create().SetName(provider.Name).SetEnabled(provider.Enabled).SetBaseURL(provider.BaseURL).SetKey(provider.Key).SetType(provider.Type).Save(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("provider.CreateProvider: creating provider: %w", err))
	}
	for _, m := range provider.Models {
		err = tx.Model.Create().SetModel(m.Model).SetAlias(m.Alias).SetProvider(pv).SetMaxTokens(int(m.MaxTokens)).SetRpm(m.Rpm).SetImage(m.Image).Exec(ctx)
		if err != nil {
			return ent.Rollback(tx, fmt.Errorf("provider.CreateProvider: creating model: %w", err))
		}
	}
	return tx.Commit()
}

func (p *ProviderServiceImpl) UpdateProvider(ctx context.Context, id int, provider Provider) error {
	p.mu.Lock()
	defer func() {
		ps, err := p.genProviders(ctx)
		if err != nil {
			p.logger.Errorw("delete provider failed", "error", err)
		} else {
			p.providers = ps
		}
		if p.syncCallback != nil {
			p.syncCallback(ctx, ps)
		}
		p.mu.Unlock()
	}()
	tx, err := p.db.Tx(ctx)
	if err != nil {
		return fmt.Errorf("provider.UpdateProvider: starting transaction: %w", err)
	}
	pv, err := tx.Provider.Query().Where(db_provider.ID(id)).Only(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("provider.UpdateProvider: querying provider: %w", err))
	}
	err = pv.Update().SetBaseURL(provider.BaseURL).SetEnabled(provider.Enabled).SetKey(provider.Key).SetType(provider.Type).Exec(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("provider.UpdateProvider: updating provider: %w", err))
	}
	_, err = tx.Model.Delete().Where(model.HasProviderWith(db_provider.ID(pv.ID))).Exec(ctx)
	if err != nil {
		return ent.Rollback(tx, fmt.Errorf("provider.UpdateProvider: deleting old models: %w", err))
	}
	for _, m := range provider.Models {
		err = tx.Model.Create().SetModel(m.Model).SetImage(m.Image).SetAlias(m.Alias).SetMaxTokens(int(m.MaxTokens)).SetRpm(m.Rpm).SetProvider(pv).Exec(ctx)
		if err != nil {
			return ent.Rollback(tx, fmt.Errorf("provider.UpdateProvider: creating model: %w", err))
		}
	}
	return tx.Commit()
}

func (p *ProviderServiceImpl) DeleteProvider(ctx context.Context, id int) error {
	p.mu.Lock()
	defer func() {
		ps, err := p.genProviders(ctx)
		if err != nil {
			p.logger.Errorw("delete provider failed", "error", err)
		} else {
			p.providers = ps
		}
		if p.syncCallback != nil {
			p.syncCallback(ctx, ps)
		}
		p.mu.Unlock()
	}()
	err := p.db.Provider.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("provider.DeleteProvider: deleting provider: %w", err)
	}
	return nil
}

func (p *ProviderServiceImpl) ListProviders(ctx context.Context) ([]Provider, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.providers, nil
}

func (p *ProviderServiceImpl) BuildProviders(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	ps, err := p.genProviders(ctx)
	if err != nil {
		return fmt.Errorf("provider.BuildProviders: generating providers: %w", err)
	}
	p.providers = ps
	return nil
}

var baseUrlMapping = map[ProviderType]string{
	ProviderTypeOpenAI:     "https://api.openai.com/v1",
	ProviderOpenRouter:     "https://openrouter.ai/api/v1",
	ProviderTypeAnthropic:  "https://api.anthropic.com/v1/",
	ProviderTypeOpenGemini: "https://generativelanguage.googleapis.com/v1beta/openai/",
}

func sameDomain(u1, u2 string) (bool, error) {
	get := func(raw string) (string, error) {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", err
		}
		return strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www."), nil
	}

	d1, err1 := get(u1)
	if err1 != nil {
		return false, err1
	}
	d2, err2 := get(u2)
	if err2 != nil {
		return false, err2
	}
	return d1 == d2, nil
}

func addProviderBaseURL(p Provider) Provider {
	// backward compatible fix
	switch p.Type {
	case "openai": // openai or openai-compatible
		p.Type = string(ProviderTypeOpenAIcompatible)
		for pt, url := range baseUrlMapping {
			match, err := sameDomain(url, p.BaseURL)
			if err != nil || !match {
				continue
			}
			p.Type = string(pt)
			break
		}
	case "gemini":
		p.Type = string(ProviderTypeOpenGemini)
	}

	switch p.Type {
	case string(ProviderTypeOpenAI):
		p.BaseURL = "https://api.openai.com/v1"
	case string(ProviderTypeOpenGemini):
		p.BaseURL = "https://generativelanguage.googleapis.com/v1beta/openai/"
	case string(ProviderTypeAnthropic):
		p.BaseURL = "https://api.anthropic.com/v1/"
	case string(ProviderOpenRouter):
		p.BaseURL = "https://openrouter.ai/api/v1"
	}
	return p
}

func (p *ProviderServiceImpl) genProviders(ctx context.Context) ([]Provider, error) {
	// add providers from config file
	providers := []Provider{}
	pm := map[string]int{}
	for i, c := range p.cfg.Providers {
		switch v := c.(type) {
		case *config.OpenAI:
			providers = append(providers, Provider{
				Name:     v.Name,
				Editable: false,
				Type:     v.Type,
				Key:      v.Key,
				BaseURL:  v.BaseURL,
				Enabled:  true,
			})
			pm[v.Name] = i
		case *config.Gemini:
			providers = append(providers, Provider{
				Name:     v.Name,
				Editable: false,
				Type:     v.Type,
				Key:      v.Key,
				Enabled:  true,
			})
			pm[v.Name] = i
		default:
			return nil, fmt.Errorf("provider.genProviders: unknown config type")
		}
	}
	for _, m := range p.cfg.Models {
		if ic, ok := pm[m.Provider]; ok {
			md := Model{
				Model:     m.Model,
				MaxTokens: m.MaxTokens,
				Alias:     m.Alias,
				Client:    m.Provider,
				Rpm:       m.RPM,
				Image:     m.Image,
				Default:   m.Default,
			}
			providers[ic].Models = append(providers[ic].Models, md)
		}
	}

	// add providers from database
	dbProviders, err := p.db.Provider.Query().Order(ent.Asc(db_provider.FieldID)).WithModels(
		func(mq *ent.ModelQuery) {
			mq.Order(ent.Asc(model.FieldModel))
		}).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("provider.genProviders: querying providers: %w", err)
	}
	for _, p := range dbProviders {
		pv := Provider{
			ID:       p.ID,
			Name:     p.Name,
			Type:     p.Type,
			Editable: true,
			Key:      p.Key,
			BaseURL:  p.BaseURL,
			Enabled:  p.Enabled,
		}
		for _, model := range p.Edges.Models {
			pv.Models = append(pv.Models, Model{
				Model:     model.Model,
				Image:     model.Image,
				Alias:     model.Alias,
				Client:    p.Name,
				MaxTokens: int64(model.MaxTokens),
				Rpm:       model.Rpm,
				Default:   model.Default,
			})
		}
		providers = append(providers, pv)
	}
	for i, p := range providers {
		providers[i] = addProviderBaseURL(p)
	}
	return providers, nil
}

func (p *ProviderServiceImpl) WithSyncCallback(callback func(ctx context.Context, providers []Provider)) {
	p.syncCallback = callback
}
