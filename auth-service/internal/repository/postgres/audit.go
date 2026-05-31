package postgres

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLogger struct {
	db *pgxpool.Pool
}

func NewAuditLogger(db *pgxpool.Pool) *AuditLogger {
	return &AuditLogger{db: db}
}

func (a *AuditLogger) Log(ctx context.Context, event string, userID *uuid.UUID, ip string, meta map[string]any) {
	var metaJSON []byte
	if meta != nil {
		b, err := json.Marshal(meta)
		if err == nil {
			metaJSON = b
		}
	}

	// fire and forget — audit failures should never block auth
	_, err := a.db.Exec(ctx, `
		INSERT INTO audit_log (user_id, event, ip_address, metadata)
		VALUES ($1, $2, $3::inet, $4)
	`, userID, event, ip, metaJSON)

	if err != nil {
		slog.Error("audit log failed", "event", event, "err", err)
	}
}