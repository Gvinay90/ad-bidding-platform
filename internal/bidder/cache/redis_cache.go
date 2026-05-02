package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Targeting struct {
	Geo, Device, Category string
}

type CampaignLite struct {
	ID            string
	BidPriceCents int64
}

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func campaignKey(campaignID string) string {
	return fmt.Sprintf("campaign:%s", campaignID)
}

func geoIdx(g string) string {
	return "idx:geo:" + g
}

func deviceIdx(d string) string {
	return "idx:device:" + d
}

func categoryIdx(c string) string {
	return "idx:category:" + c
}

const activeIdx = "idx:status:active"

// upsert writes to campaign cache and updates indexes
func (c *Cache) Upsert(ctx context.Context, id string, bidPrice int64, geo, device, category, status string) error {
	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, campaignKey(id), map[string]interface{}{
		"bid_price_cents": bidPrice,
		"geo":             geo,
		"device":          device,
		"category":        category,
		"status":          status,
	})
	pipe.SAdd(ctx, geoIdx(geo), id)
	pipe.SAdd(ctx, deviceIdx(device), id)
	pipe.SAdd(ctx, categoryIdx(category), id)
	if status == "active" {
		pipe.SAdd(ctx, activeIdx, id)
	} else {
		pipe.SRem(ctx, activeIdx, id)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) Delete(ctx context.Context, id string) error {
	return c.rdb.Del(ctx, campaignKey(id)).Err()
}

func (c *Cache) GetCampaignLite(ctx context.Context, targeting Targeting) ([]CampaignLite, error) {
	keys := []string{
		activeIdx,
		geoIdx(targeting.Geo),
		deviceIdx(targeting.Device),
		categoryIdx(targeting.Category),
	}
	ids, err := c.rdb.SInter(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	pipe := c.rdb.Pipeline()
	cmds := make([]*redis.SliceCmd, 0, len(ids))
	for _, id := range ids {
		cmds = append(cmds, pipe.HMGet(ctx, campaignKey(id), "id", "bid_price"))
	}
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}
	out := make([]CampaignLite, 0, len(cmds))
	for _, cmd := range cmds {
		vals, err := cmd.Result()
		if err != nil || len(vals) < 2 || vals[0] == nil {
			continue
		}
		id, _ := vals[0].(string)
		priceStr, _ := vals[1].(string)
		price, perr := strconv.ParseInt(priceStr, 10, 64)
		if perr != nil {
			continue
		}
		out = append(out, CampaignLite{ID: id, BidPriceCents: price})
	}
	return out, nil

}

func (c *Cache) Stats(ctx context.Context) (string, error) {
	n, err := c.rdb.SCard(ctx, activeIdx).Result()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("active: %d", n), nil
}
