package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/port"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func New(ctx context.Context) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &Redis{client: client}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) Get(ctx context.Context, key string) (*domain.Forecast, error) {
	res, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result domain.Forecast
	if err = json.Unmarshal([]byte(res), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *Redis) Set(ctx context.Context, key string, value *domain.Forecast, ttl time.Duration) error {
	valueSerialized, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, valueSerialized, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *Redis) AddGeoData(ctx context.Context, geoKey, locationName string, longitude, latitude float64) error {
	return r.client.GeoAdd(ctx, geoKey, &redis.GeoLocation{
		Name:      locationName,
		Longitude: longitude,
		Latitude:  latitude,
	}).Err()
}

func (r *Redis) DeleteGeoData(ctx context.Context, geoKey, locationName string) error {
	return r.client.ZRem(ctx, geoKey, locationName).Err()
}

func (r *Redis) FindKeyWithinRadiusWithUpsert(ctx context.Context, geoKey, locationName string, longitude, latitude, radius float64) (string, error) {
	script := redis.NewScript(`
		local existing = redis.call(
		  "GEOSEARCH",
		  KEYS[1],
		  "FROMLONLAT", ARGV[1], ARGV[2],
		  "BYRADIUS", ARGV[4], "m",
		  "COUNT", 1
		)

		if #existing > 0 then
		  return {0, existing[1]}
		end

		redis.call("GEOADD", KEYS[1], ARGV[1], ARGV[2], ARGV[3])
		return {1, ARGV[3]}
		`,
	)

	res, err := script.Run(
		ctx,
		r.client,
		[]string{geoKey},
		longitude,
		latitude,
		locationName,
		radius,
	).Slice()

	if err != nil {
		return "", err
	}

	member := res[1].(string)
	return member, nil
}

func (r *Redis) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, "lock:"+key, 1, ttl).Result()
}

func (r *Redis) Unlock(ctx context.Context, key string) error {
	return r.Delete(ctx, "lock:"+key)
}

func (r *Redis) Subscribe(ctx context.Context, channelName string) (port.CancelFunc, port.NextMessageFunc) {
	s := r.client.Subscribe(ctx, "channel:" + channelName)
	ch := s.Channel()
	return s.Close, func() (string, bool) {
		msg, ok := <-ch
		if !ok || msg == nil {
			return "", false
		}
		return msg.Payload, ok
	}
}


func (r *Redis) Publish(ctx context.Context, channelName string, message any) error {
	return r.client.Publish(ctx, "channel:" + channelName, message).Err()
}
