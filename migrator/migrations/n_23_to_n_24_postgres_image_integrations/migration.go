// Code generated by pg-bindings generator. DO NOT EDIT.
package n23ton24

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/migrator/migrations"
	frozenSchema "github.com/stackrox/rox/migrator/migrations/frozenschema/v73"
	"github.com/stackrox/rox/migrator/migrations/loghelper"
	legacy "github.com/stackrox/rox/migrator/migrations/n_23_to_n_24_postgres_image_integrations/legacy"
	pgStore "github.com/stackrox/rox/migrator/migrations/n_23_to_n_24_postgres_image_integrations/postgres"
	"github.com/stackrox/rox/migrator/types"
	pkgMigrations "github.com/stackrox/rox/pkg/migrations"
	"github.com/stackrox/rox/pkg/postgres/pgutils"
	"github.com/stackrox/rox/pkg/sac"
	"gorm.io/gorm"
)

var (
	migration = types.Migration{
		StartingSeqNum: pkgMigrations.BasePostgresDBVersionSeqNum() + 23,
		VersionAfter:   &storage.Version{SeqNum: int32(pkgMigrations.BasePostgresDBVersionSeqNum()) + 24},
		Run: func(databases *types.Databases) error {
			legacyStore := legacy.New(databases.BoltDB)
			if err := move(databases.GormDB, databases.PostgresDB, legacyStore); err != nil {
				return errors.Wrap(err,
					"moving image_integrations from rocksdb to postgres")
			}
			return nil
		},
	}
	batchSize = 10000
	schema    = frozenSchema.ImageIntegrationsSchema
	log       = loghelper.LogWrapper{}
)

func move(gormDB *gorm.DB, postgresDB *pgxpool.Pool, legacyStore legacy.Store) error {
	ctx := sac.WithAllAccess(context.Background())
	store := pgStore.New(postgresDB)
	pgutils.CreateTableFromModel(context.Background(), gormDB, frozenSchema.CreateTableImageIntegrationsStmt)
	imageIntegrations, err := legacyStore.GetAll(ctx)
	if err != nil {
		return err
	}
	if len(imageIntegrations) > 0 {
		if err = store.UpsertMany(ctx, imageIntegrations); err != nil {
			log.WriteToStderrf("failed to persist image_integrations to store %v", err)
			return err
		}
	}
	return nil
}

func init() {
	migrations.MustRegisterMigration(migration)
}
