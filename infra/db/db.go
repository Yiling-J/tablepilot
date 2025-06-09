package db

import (
	"bytes"
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/Yiling-J/tablepilot/config"
	"github.com/Yiling-J/tablepilot/ent"
	"github.com/Yiling-J/tablepilot/ent/hook"

	"github.com/spf13/cast"
	"github.com/sqids/sqids-go"
	"go.uber.org/zap"
	"modernc.org/sqlite"
)

func init() {
	sql.Register("sqlite3", &sqlite.Driver{})
}

const (
	nanoidAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	nanoidLength   = 6
)

type NanoidMutation interface {
	Nanoid() (r string, exists bool)
	ID() (id int, exists bool)
}

type Updater interface {
	Exec(context.Context) error
}

func NanoIDHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		sq, _ := sqids.New(sqids.Options{Alphabet: nanoidAlphabet, MinLength: 6})
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			v, err := next.Mutate(ctx, m)
			if err != nil {
				return v, err
			}
			var nanoid string
			if n, ok := m.(NanoidMutation); ok {
				id, _ := n.ID()
				nanoid, _ = sq.Encode([]uint64{cast.ToUint64(id)})
			} else {
				return v, err
			}

			var updater Updater
			switch vt := v.(type) {
			case *ent.TableMeta:
				vt.Nanoid = nanoid
				updater = vt.Update().SetNanoid(nanoid)
			case *ent.TableColumn:
				vt.Nanoid = nanoid
				updater = vt.Update().SetNanoid(nanoid)
			case *ent.TableRow:
				vt.Nanoid = nanoid
				updater = vt.Update().SetNanoid(nanoid)
			case *ent.Workflow:
				vt.Nanoid = nanoid
				updater = vt.Update().SetNanoid(nanoid)
			case *ent.Dataset:
				vt.Nanoid = nanoid
				updater = vt.Update().SetNanoid(nanoid)
			default:
				return v, err
			}
			err = updater.Exec(ctx)
			return v, err
		})
	}
}

func NewDB(config *config.Config, logger *zap.SugaredLogger) (*ent.Client, error) {
	logger.Debugw(
		"conntect to database", "driver", config.Database.Driver, "dsn", config.Database.DSN,
	)
	client, err := ent.Open(config.Database.Driver, config.Database.DSN)
	if err != nil {
		return nil, err
	}
	client.Use(hook.On(NanoIDHook(), ent.OpCreate))
	logger.Debug("database connected")
	logger.Debug("stating migration")
	bf := bytes.NewBuffer([]byte{})
	err = client.Schema.WriteTo(context.TODO(), bf)
	if err != nil {
		return nil, err
	}
	sqls := bf.String()
	cts := strings.Count(sqls, "CREATE TABLE")
	if cts > 3 {
		config.ShouldCreateExampleTable = true
	}
	if err := client.Schema.Create(context.TODO()); err != nil {
		return nil, err
	}
	logger.Debug("migration done")
	return client, nil
}

func NewTestDB() *ent.Client {
	client, err := ent.Open("sqlite3", ":memory:?_pragma=foreign_keys(1)")
	if err != nil {
		log.Fatalf("failed connecting to sqlite: %v", err)
	}
	client.Use(hook.On(NanoIDHook(), ent.OpCreate))
	if err := client.Schema.Create(context.TODO()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	return client
}
