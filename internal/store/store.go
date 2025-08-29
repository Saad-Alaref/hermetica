package store

import (
    "context"
    "database/sql"
    _ "modernc.org/sqlite"
)

type DB struct {
    sql *sql.DB
}

func Open(path string) (*DB, error) {
    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, err
    }
    if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
        return nil, err
    }
    s := &DB{sql: db}
    if err := s.migrate(context.Background()); err != nil {
        return nil, err
    }
    return s, nil
}

func (d *DB) Close() error { return d.sql.Close() }

func (d *DB) migrate(ctx context.Context) error {
    stmts := []string{
        `CREATE TABLE IF NOT EXISTS assets (
            id TEXT PRIMARY KEY,
            domain TEXT,
            subdomain TEXT,
            fqdn TEXT,
            ip TEXT,
            rrtype TEXT,
            first_seen TIMESTAMP,
            last_seen TIMESTAMP
        );`,
        `CREATE INDEX IF NOT EXISTS idx_assets_fqdn ON assets(fqdn);`,
        `CREATE INDEX IF NOT EXISTS idx_assets_ip ON assets(ip);`,
        `CREATE TABLE IF NOT EXISTS services (
            asset_id TEXT,
            ip TEXT,
            port INTEGER,
            proto TEXT,
            is_web BOOLEAN,
            PRIMARY KEY (ip, port, proto)
        );`,
        `CREATE TABLE IF NOT EXISTS webtargets (
            service_id TEXT,
            input_host TEXT,
            sni_mode TEXT,
            url TEXT,
            status INTEGER,
            title TEXT,
            final_url TEXT,
            tls_issuer TEXT,
            cdn_hint TEXT,
            tech TEXT,
            body_hash TEXT,
            page_group TEXT,
            body_path TEXT,
            shot_path TEXT,
            PRIMARY KEY (service_id, sni_mode, input_host, url)
        );`,
        `CREATE TABLE IF NOT EXISTS discovery (
            source TEXT,
            hostname TEXT,
            in_scope BOOLEAN,
            note TEXT,
            seen_at TIMESTAMP
        );`,
    }
    for _, s := range stmts {
        if _, err := d.sql.ExecContext(ctx, s); err != nil {
            return err
        }
    }
    return nil
}

