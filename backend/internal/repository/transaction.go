package repository

import (
	"context"

	"gorm.io/gorm"
)

type TransactionRunner interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type GormTransactionRunner struct {
	db *gorm.DB
}

type gormTransactionContextKey struct{}

func NewGormTransactionRunner(db *gorm.DB) *GormTransactionRunner {
	return &GormTransactionRunner{db: db}
}

func (r *GormTransactionRunner) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if hasGormTransaction(ctx) {
		return fn(ctx)
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, gormTransactionContextKey{}, tx))
	})
}

func hasGormTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(gormTransactionContextKey{}).(*gorm.DB)
	return ok
}

func gormDB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(gormTransactionContextKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}
