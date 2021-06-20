package db

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"go.uber.org/zap"
)

const (
	SlowThreshold time.Duration = 500 * time.Millisecond
)

type DBLogger struct {
	logger *zap.Logger
}

func NewLogger(logger *zap.Logger) pg.QueryHook {
	return &DBLogger{logger}
}

func (l *DBLogger) BeforeQuery(ctx context.Context, _ *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (l *DBLogger) AfterQuery(ctx context.Context, q *pg.QueryEvent) error {
	elapsed := time.Since(q.StartTime)

	err := q.Err
	sql, _ := q.FormattedQuery() // err is always nil?

	if err != nil {
		l.logger.Warn("trace", zap.Error(err), zap.Duration("elapsed", elapsed), zap.String("sql", string(sql)))
	} else if elapsed > SlowThreshold {
		l.logger.Warn("trace", zap.Duration("elapsed", elapsed), zap.String("sql", string(sql)))
	} else {
		l.logger.Info("trace", zap.Duration("elapsed", elapsed), zap.String("sql", string(sql)))
	}

	return nil
}
