package service_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mumtozvalijonov/weather/internal/core/domain"
	"github.com/mumtozvalijonov/weather/internal/core/port"
	"github.com/mumtozvalijonov/weather/internal/core/service"
)

var nextID atomic.Uint64

type fakePubSub struct {
	name        string
	pubChannel  chan string
	subChannels map[uint64]chan string
	running     atomic.Bool
}

func newFakePubSub(name string) *fakePubSub {
	return &fakePubSub{
		name:        name,
		pubChannel:  make(chan string, 1),
		subChannels: make(map[uint64]chan string),
	}
}

func (ps *fakePubSub) run() {
	if !ps.running.CompareAndSwap(false, true) {
		return
	}

	for {
		msg := <-ps.pubChannel
		for _, ch := range ps.subChannels {
			ch <- msg // TODO: account for when channel is closed
		}
	}
}

type fakeWeatherRepository struct {
	mu      sync.Mutex
	data    map[string]*domain.Forecast
	geo     map[string]map[string]fakeGeoLocation
	locks   map[string]*sync.Mutex
	pubSubs map[string]*fakePubSub
}

type fakeGeoLocation struct {
	longitude float64
	latitude  float64
}

func newFakeWeatherRepository() *fakeWeatherRepository {
	return &fakeWeatherRepository{
		data: make(map[string]*domain.Forecast),
		geo:  make(map[string]map[string]fakeGeoLocation),
	}
}

func (f *fakeWeatherRepository) Get(ctx context.Context, key string) (*domain.Forecast, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	value, ok := f.data[key]
	if !ok {
		return nil, service.ErrForecastUnavailable
	}
	return value, nil
}

func (f *fakeWeatherRepository) Set(ctx context.Context, key string, value *domain.Forecast, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.data[key] = value
	return nil
}

func (f *fakeWeatherRepository) Delete(ctx context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.data, key)
	return nil
}

func (f *fakeWeatherRepository) AddGeoData(ctx context.Context, geoKey, locationName string, longitude, latitude float64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.addGeoData(geoKey, locationName, longitude, latitude)
	return nil
}

func (f *fakeWeatherRepository) addGeoData(geoKey, locationName string, longitude, latitude float64) {
	if f.geo[geoKey] == nil {
		f.geo[geoKey] = map[string]fakeGeoLocation{}
	}

	f.geo[geoKey][locationName] = fakeGeoLocation{
		longitude: longitude,
		latitude:  latitude,
	}
}

func (f *fakeWeatherRepository) DeleteGeoData(ctx context.Context, geoKey, locationName string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.geo[geoKey], locationName)
	return nil
}

func (f *fakeWeatherRepository) findKeyWithinRadius(geoKey string, longitude, latitude, radius float64) string {
	for locationName, location := range f.geo[geoKey] {
		distance := distanceMeters(latitude, longitude, location.latitude, location.longitude)
		if distance <= radius {
			return locationName
		}
	}

	return ""
}

func (f *fakeWeatherRepository) FindKeyWithinRadiusWithUpsert(ctx context.Context, geoKey, locationName string, longitude, latitude, radius float64) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	existingLocationName := f.findKeyWithinRadius(geoKey, longitude, latitude, radius)
	if existingLocationName != "" {
		return existingLocationName, nil
	}

	f.addGeoData(geoKey, locationName, longitude, latitude)
	return locationName, nil
}

func (f *fakeWeatherRepository) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	clear(f.data)
	clear(f.geo)
	return nil
}

func (f *fakeWeatherRepository) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.locks == nil {
		f.locks = make(map[string]*sync.Mutex)
	}
	if _, ok := f.locks[key]; !ok {
		f.locks[key] = &sync.Mutex{}
	}
	ok := f.locks[key].TryLock()
	return ok, nil
}

func (f *fakeWeatherRepository) Unlock(ctx context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.locks == nil {
		return nil
	}

	if c, ok := f.locks[key]; ok {
		c.Unlock()
		delete(f.locks, key)
		return nil
	}
	return nil
}

func (f *fakeWeatherRepository) Subscribe(ctx context.Context, channelName string) (port.CancelFunc, port.NextMessageFunc) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.pubSubs == nil {
		f.pubSubs = make(map[string]*fakePubSub)
	}

	if _, ok := f.pubSubs[channelName]; !ok {
		f.pubSubs[channelName] = newFakePubSub(channelName)
	}

	pubSub := f.pubSubs[channelName]
	subCh := make(chan string)
	subId := nextID.Add(1)
	pubSub.subChannels[subId] = subCh

	chanCtx, cancel := context.WithCancel(ctx)

	cancelFunc := func() error {
		delete(pubSub.subChannels, subId)
		close(subCh)
		cancel()
		return nil
	}

	nextMessageFunc := func() (string, bool) {
		select {
		case msg, ok := <-subCh:
			if ok {
				return msg, true
			}
		case <-chanCtx.Done():
		}
		return "", false
	}

	return cancelFunc, nextMessageFunc
}

func (f *fakeWeatherRepository) Publish(ctx context.Context, channelName string, message any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.pubSubs == nil {
		f.pubSubs = make(map[string]*fakePubSub)
	}

	if _, ok := f.pubSubs[channelName]; !ok {
		f.pubSubs[channelName] = newFakePubSub(channelName)
	}

	go f.pubSubs[channelName].run()

	select {
	case f.pubSubs[channelName].pubChannel <- fmt.Sprintf("%v", message):
	case <-ctx.Done():
	}
	return nil
}
