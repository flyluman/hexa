package ipfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"portfolio/internal/core/domain"
	"portfolio/pkg/keylock"
	"time"
)

type ipWhoisAdapter struct {
	client *http.Client
	kl     *keylock.KeyLock[*domain.MetaIP]
}

func NewIPWhoisAdapter() *ipWhoisAdapter {
	return &ipWhoisAdapter{
		client: &http.Client{Timeout: 5 * time.Second},
		kl:     keylock.NewKeyLock[*domain.MetaIP](),
	}
}

func (a *ipWhoisAdapter) FetchIPInfo(ctx context.Context, ip string) (*domain.MetaIP, error) {
	return a.kl.Read(ip, func() (*domain.MetaIP, error) {
		url := fmt.Sprintf("http://ipwhois.app/json/%s?objects=ip,isp,city,country", ip)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("ip:%s, error:%w", ip, err)
		}

		res, err := a.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ip:%s, error:%w", ip, err)
		}
		defer res.Body.Close()

		var metaIP domain.MetaIP
		if err := json.NewDecoder(res.Body).Decode(&metaIP); err != nil {
			return nil, fmt.Errorf("ip:%s, error:%w", ip, err)
		}

		return &metaIP, nil
	})
}
