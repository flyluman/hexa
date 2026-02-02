package repo

// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"fmt"
// 	"portfolio/internal/core/domain"
// 	"portfolio/internal/util"
// 	"time"

// 	_ "github.com/mattn/go-sqlite3"
// )

// type SQLiteRepo struct {
// 	client *sql.DB
// }

// func NewSQLiteRepo() (*SQLiteRepo, error) {
// 	client, err := sql.Open("sqlite3", "./data/portfolio.db")
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := client.Ping(); err != nil {
// 		return nil, err
// 	}

// 	client.SetMaxOpenConns(1)
// 	client.SetMaxIdleConns(1)

// 	r := &SQLiteRepo{
// 		client: client,
// 	}

// 	if err := r.migrate(); err != nil {
// 		return nil, err
// 	}

// 	return r, nil
// }

// func (r *SQLiteRepo) migrate() error {
// 	query := `CREATE TABLE IF NOT EXISTS log (
// 		id TEXT PRIMARY KEY,
// 		ip TEXT, isp TEXT, city TEXT, country TEXT,
// 		date TEXT, path TEXT, useragent TEXT
// 	);
// 		CREATE TABLE IF NOT EXISTS msg (
// 		id TEXT PRIMARY KEY,
// 		name TEXT, email TEXT, msg TEXT
// 	);`

// 	if _, err := r.client.Exec(query); err != nil {
// 		return fmt.Errorf("query:%s, error:%w", query, err)
// 	}

// 	return nil
// }

// func (r *SQLiteRepo) SaveMessage(ctx context.Context, msg *domain.Message) error {
// 	query := "INSERT INTO msg (id, name, email, msg) VALUES (?, ?, ?, ?)"
// 	if _, err := r.client.ExecContext(ctx, query, ctx.Value("request_id"), msg.Name, msg.Email, msg.Text); err != nil {
// 		return fmt.Errorf("query:%s, error:%w", query, err)
// 	}

// 	return nil
// }

// func (r *SQLiteRepo) SaveMeta(ctx context.Context, metaReq *domain.MetaReq, metaIP *domain.MetaIP) error {
// 	query := "INSERT INTO log (id, ip, isp, city, country, date, path, useragent) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
// 	if _, err := r.client.ExecContext(ctx, query, metaReq.RequestID, metaReq.IP, metaIP.ISP, metaIP.City, metaIP.Country, util.Now().Format(time.RFC3339), metaReq.Path, metaReq.Useragent); err != nil {
// 		return fmt.Errorf("query:%s, error:%w", query, err)
// 	}

// 	return nil
// }

// func (r *SQLiteRepo) QueryLog(ctx context.Context) ([]*domain.Log, error) {
// 	query := "SELECT id, ip, isp, city, country, date, path, useragent FROM log ORDER BY date DESC LIMIT 50"

// 	res, err := r.client.QueryContext(ctx, query)
// 	if err != nil {
// 		return nil, fmt.Errorf("query:%s, error:%w", query, err)
// 	}
// 	defer res.Close()

// 	ret := make([]*domain.Log, 0, 50)

// 	for res.Next() {
// 		var tmp domain.Log
// 		if localErr := res.Scan(&tmp.ID, &tmp.IP, &tmp.ISP, &tmp.City, &tmp.Country, &tmp.Date, &tmp.Path, &tmp.Useragent); localErr != nil {
// 			err = errors.Join(err, localErr)
// 		}
// 		ret = append(ret, &tmp)
// 	}

// 	if err != nil {
// 		return ret, fmt.Errorf("query:%s, error:%w", query, err)
// 	}

// 	return ret, nil
// }
