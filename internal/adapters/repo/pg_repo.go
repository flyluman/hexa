package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"portfolio/internal/core/domain"
	"portfolio/internal/util"

	_ "github.com/lib/pq"
)

type PostgresRepo struct {
	client *sql.DB
}

func NewPostgresRepo(host, port, user, pass, dbName string) (*PostgresRepo, error) {
	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbName)

	client, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	r := &PostgresRepo{
		client: client,
	}

	if err := r.migrate(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *PostgresRepo) migrate() error {
	query := `CREATE TABLE IF NOT EXISTS logs (
		id TEXT PRIMARY KEY,
		ip TEXT, isp TEXT, city TEXT, country TEXT,
		date TIMESTAMP, path TEXT, useragent TEXT
	);
		CREATE TABLE IF NOT EXISTS msg (
		id TEXT PRIMARY KEY,
		name TEXT, email TEXT, msg TEXT
	);`

	if _, err := r.client.Exec(query); err != nil {
		return fmt.Errorf("query:%s, error:%w", query, err)
	}

	return nil
}

func (r *PostgresRepo) SaveMessage(ctx context.Context, msg *domain.Message) error {
	query := "INSERT INTO msg (id, name, email, msg) VALUES ($1, $2, $3, $4)"
	if _, err := r.client.ExecContext(ctx, query, msg.RequestID, msg.Name, msg.Email, msg.Text); err != nil {
		return fmt.Errorf("query:%s, error:%w", query, err)
	}

	return nil
}

func (r *PostgresRepo) SaveMeta(ctx context.Context, metaReq *domain.MetaReq, metaIP *domain.MetaIP) error {
	query := "INSERT INTO logs (id, ip, isp, city, country, date, path, useragent) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	if _, err := r.client.ExecContext(ctx, query, metaReq.RequestID, metaReq.IP, metaIP.ISP, metaIP.City, metaIP.Country, util.Now(), metaReq.Path, metaReq.Useragent); err != nil {
		return fmt.Errorf("query:%s, error:%w", query, err)
	}

	return nil
}

func (r *PostgresRepo) QueryLog(ctx context.Context) ([]*domain.Log, error) {
	query := "SELECT id, ip, isp, city, country, date, path, useragent FROM logs ORDER BY date DESC LIMIT 50"

	res, err := r.client.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query:%s, error:%w", query, err)
	}
	defer res.Close()

	ret := make([]*domain.Log, 0, 50)

	for res.Next() {
		var tmp domain.Log
		if localErr := res.Scan(&tmp.ID, &tmp.IP, &tmp.ISP, &tmp.City, &tmp.Country, &tmp.Date, &tmp.Path, &tmp.Useragent); localErr != nil {
			err = errors.Join(err, localErr)
		}
		ret = append(ret, &tmp)
	}

	if err != nil {
		return ret, fmt.Errorf("query:%s, error:%w", query, err)
	}

	return ret, nil
}
