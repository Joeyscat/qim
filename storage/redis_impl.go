package storage

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/wire/pkt"
)

const (
	LocationExpired = time.Hour * 48
)

type RedisStorage struct {
	cli *redis.Client
}

func NewRedisStorage(cli *redis.Client) qim.SessionStorage {
	return &RedisStorage{
		cli: cli,
	}
}

// Add implements qim.SessionStorage
func (*RedisStorage) Add(session *pkt.Session) error {
	panic("unimplemented")
}

// Delete implements qim.SessionStorage
func (*RedisStorage) Delete(account string, channelID string) error {
	panic("unimplemented")
}

// Get implements qim.SessionStorage
func (*RedisStorage) Get(channelID string) (*pkt.Session, error) {
	panic("unimplemented")
}

// GetLocation implements qim.SessionStorage
func (*RedisStorage) GetLocation(account string, device string) (*qim.Location, error) {
	panic("unimplemented")
}

// GetLocations implements qim.SessionStorage
func (*RedisStorage) GetLocations(account ...string) ([]*qim.Location, error) {
	panic("unimplemented")
}
