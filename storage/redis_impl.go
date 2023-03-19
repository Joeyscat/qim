package storage

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joeyscat/qim"
	"github.com/joeyscat/qim/wire/pkt"
	"google.golang.org/protobuf/proto"
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
func (s *RedisStorage) Add(session *pkt.Session) error {
	// save qim.Location
	loc := qim.Location{
		ChannelID: session.GetChannelId(),
		GateID:    session.GetGateId(),
	}
	locKey := KeyLocation(session.GetAccount(), session.GetDevice())
	err := s.cli.Set(s.cli.Context(), locKey, loc.Bytes(), LocationExpired).Err()
	if err != nil {
		return err
	}

	// save session
	snKey := KeySession(session.GetChannelId())
	buf, _ := proto.Marshal(session)
	return s.cli.Set(s.cli.Context(), snKey, buf, LocationExpired).Err()
}

// Delete implements qim.SessionStorage
func (s *RedisStorage) Delete(account string, channelID string) error {
	locKey := KeyLocation(account, "")
	err := s.cli.Del(s.cli.Context(), locKey).Err()
	if err != nil {
		return err
	}

	snKey := KeySession(channelID)
	return s.cli.Del(s.cli.Context(), snKey).Err()
}

// Get implements qim.SessionStorage
func (s *RedisStorage) Get(channelID string) (*pkt.Session, error) {
	snKey := KeySession(channelID)
	bts, err := s.cli.Get(s.cli.Context(), snKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, qim.ErrSessionNil
		}
		return nil, err
	}

	var session pkt.Session
	_ = proto.Unmarshal(bts, &session)
	return &session, nil
}

// GetLocation implements qim.SessionStorage
func (s *RedisStorage) GetLocation(account string, device string) (*qim.Location, error) {
	key := KeyLocation(account, device)
	bts, err := s.cli.Get(s.cli.Context(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, qim.ErrSessionNil
		}
		return nil, err
	}

	var loc qim.Location
	_ = loc.Unmarshal(bts)
	return &loc, nil
}

// GetLocations implements qim.SessionStorage
func (s *RedisStorage) GetLocations(account ...string) ([]*qim.Location, error) {
	keys := KeyLocations(account...)
	list, err := s.cli.MGet(s.cli.Context(), keys...).Result()
	if err != nil {
		return nil, err
	}

	var result = make([]*qim.Location, 0)
	for _, v := range list {
		if v == nil {
			continue
		}

		var loc qim.Location
		_ = loc.Unmarshal([]byte(v.(string)))
		result = append(result, &loc)
	}

	return result, nil
}

func KeyLocation(account, device string) string {
	if device == "" {
		return fmt.Sprintf("login:loc:%s", account)
	}
	return fmt.Sprintf("login:loc:%s:%s", account, device)
}

func KeyLocations(accounts ...string) []string {
	arr := make([]string, len(accounts))
	for i, account := range accounts {
		arr[i] = KeyLocation(account, "")
	}
	return arr
}

func KeySession(channelID string) string {
	return fmt.Sprintf("login:sn:%s", channelID)
}
