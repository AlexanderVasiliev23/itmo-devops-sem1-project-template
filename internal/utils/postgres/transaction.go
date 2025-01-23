package postgres

import (
	"context"
	"database/sql"
)

const contextTxKey = "tx"

type TransactionProvider struct {
	db *sql.DB
}

func NewTransactionProvider(db *sql.DB) *TransactionProvider {
	return &TransactionProvider{
		db: db,
	}
}

func (p *TransactionProvider) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, contextTxKey, tx)

	if err := fn(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

func (p *TransactionProvider) GetDBTransaction(ctx context.Context) (*sql.Tx, error) {
	if ctx.Value(contextTxKey) == nil {
		return p.db.Begin()
	}

	return ctx.Value(contextTxKey).(*sql.Tx), nil
}
