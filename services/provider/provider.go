package provider

import (
	"context"
	"errors"
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
		return ent.Rollback(tx, err)
	}
	pv, err := tx.Provider.Create().SetName(provider.Name).SetBaseURL(provider.BaseURL).SetKey(provider.Key).SetType(db_provider.Type(provider.Type)).Save(ctx)
	if err != nil {
		return ent.Rollback(tx, err)
	}
	for _, m := range provider.Models {
		err = tx.Model.Create().SetModel(m.Model).SetAlias(m.Alias).SetProvider(pv).SetMaxTokens(int(m.MaxTokens)).SetRpm(m.Rpm).SetImage(m.Image).Exec(ctx)
		if err != nil {
			return ent.Rollback(tx, err)
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
		return ent.Rollback(tx, err)
	}
	pv, err := tx.Provider.Query().Where(db_provider.ID(id)).Only(ctx)
	if err != nil {
		return ent.Rollback(tx, err)
	}
	err = pv.Update().SetBaseURL(provider.BaseURL).SetKey(provider.Key).SetType(db_provider.Type(provider.Type)).Exec(ctx)
	if err != nil {
		return ent.Rollback(tx, err)
	}
	_, err = tx.Model.Delete().Where(model.HasProviderWith(db_provider.ID(pv.ID))).Exec(ctx)
	if err != nil {
		return ent.Rollback(tx, err)
	}
	for _, m := range provider.Models {
		err = tx.Model.Create().SetModel(m.Model).SetImage(m.Image).SetAlias(m.Alias).SetMaxTokens(int(m.MaxTokens)).SetRpm(m.Rpm).SetProvider(pv).Exec(ctx)
		if err != nil {
			return ent.Rollback(tx, err)
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
	return p.db.Provider.DeleteOneID(id).Exec(ctx)
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
		return err
	}
	p.providers = ps
	return nil
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
			})
			pm[v.Name] = i
		case *config.Gemini:
			providers = append(providers, Provider{
				Name:     v.Name,
				Editable: false,
				Type:     v.Type,
				Key:      v.Key,
			})
			pm[v.Name] = i
		default:
			return nil, errors.New("unknown config type")
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
	dbProviders, err := p.db.Provider.Query().Order(ent.Asc(db_provider.FieldName)).WithModels(
		func(mq *ent.ModelQuery) {
			mq.Order(ent.Asc(model.FieldModel))
		}).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range dbProviders {
		pv := Provider{
			ID:       p.ID,
			Name:     p.Name,
			Type:     p.Type.String(),
			Editable: true,
			Key:      p.Key,
			BaseURL:  p.BaseURL,
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
	return providers, nil
}

func (p *ProviderServiceImpl) WithSyncCallback(callback func(ctx context.Context, providers []Provider)) {
	p.syncCallback = callback
}
