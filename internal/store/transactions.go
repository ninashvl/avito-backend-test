package store

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type txCtxKey struct{}

// TxFromContext returns a Tx stored inside a context, or nil if there isn't one.
func TxFromContext(ctx context.Context) *sqlx.Tx {
	tx, _ := ctx.Value(txCtxKey{}).(*sqlx.Tx)
	return tx
}

// NewTxContext returns a new context with the given Tx attached.
func NewTxContext(parent context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(parent, txCtxKey{}, tx)
}

type Transactor interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) (err error)
}

type sqlxTransactor struct {
	db *sqlx.DB
}

func NewTransactor(db *sqlx.DB) Transactor {
	return &sqlxTransactor{
		db: db,
	}
}

func (t *sqlxTransactor) RunInTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx := TxFromContext(ctx)
	if tx == nil {
		tx, err = t.db.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("creating transaction: %v", err)
		}
	}

	defer func() {
		if p := recover(); p != nil {
			// a panic occurred, rollback and repanic
			// nolint: errcheck
			tx.Rollback()
			panic(p)
		} else if err != nil {
			// something went wrong, rollback
			// nolint: errcheck
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	err = fn(NewTxContext(ctx, tx))
	return err
}

type Executor interface {
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	squirrel.StdSqlCtx
}
