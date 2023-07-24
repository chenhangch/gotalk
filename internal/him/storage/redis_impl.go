package storage

import (
	"fmt"
	"log"
	"time"

	"github.com/chang144/gotalk/internal/him"
	"github.com/chang144/gotalk/internal/him/wire/pkt"
	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const LocationExpired = time.Hour * 2

// InitRedis return a redis instance
func InitRedis(addr string, pass string) (*redis.Client, error) {
	redisdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	})

	_, err := redisdb.Ping().Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return redisdb, nil
}

// InitFailoverRedis init redis with sentinels
func InitFailoverRedis(masterName string, sentinelAddrs []string, password string, timeout time.Duration) (*redis.Client, error) {
	redisdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      password,
		DialTimeout:   time.Second * 5,
		ReadTimeout:   timeout,
		WriteTimeout:  timeout,
	})

	_, err := redisdb.Ping().Result()
	if err != nil {
		logrus.Warn(err)
	}
	return redisdb, nil
}

type RedisStorage struct {
	cli *redis.Client
}

func NewRedisStorage(cli *redis.Client) *RedisStorage {
	return &RedisStorage{cli}
}

func (r *RedisStorage) Delete(account string, channelId string) error {
	locKey := KeyLocation(account, "")
	err := r.cli.Del(locKey).Err()
	if err != nil {
		return err
	}

	snKey := KeySession(channelId)
	err = r.cli.Del(snKey).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get GetByID to get session by sessionID
func (r *RedisStorage) Get(channelId string) (*pkt.Session, error) {
	snKey := KeySession(channelId)
	bytes, err := r.cli.Get(snKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, him.ErrSessionNil
		}
		return nil, err
	}
	var session *pkt.Session
	err = proto.Unmarshal(bytes, session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetLocations 批量读取位置信息
func (r *RedisStorage) GetLocations(account ...string) ([]*him.Location, error) {
	keys := KeyLocations(account...)
	list, err := r.cli.MGet(keys...).Result()
	if err != nil {
		return nil, err
	}
	result := make([]*him.Location, 0)
	for _, l := range list {
		if l == nil {
			continue
		}
		var loc him.Location
		err := loc.Unmarshal([]byte(l.(string)))
		if err != nil {
			return nil, err
		}
		result = append(result, &loc)
	}
	if len(result) == 0 {
		return nil, him.ErrSessionNil
	}
	return result, nil
}

func (r *RedisStorage) GetLocation(account string, device string) (*him.Location, error) {
	key := KeyLocation(account, device)
	bytes, err := r.cli.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, him.ErrSessionNil
		}
		return nil, err
	}
	var loc him.Location
	_ = loc.Unmarshal(bytes)
	return &loc, err
}

// Add 添加会话
func (r *RedisStorage) Add(session *pkt.Session) error {
	// 保存 him.location
	loc := him.Location{
		ChannelId: session.ChannelId,
		GateId:    session.GateId,
	}
	locKey := KeyLocation(session.Account, "")
	err := r.cli.Set(locKey, loc.Bytes(), LocationExpired).Err()
	if err != nil {
		return err
	}
	// save session
	snKey := KeySession(session.ChannelId)
	buf, err := proto.Marshal(session)
	if err != nil {
		return err
	}
	err = r.cli.Set(snKey, buf, LocationExpired).Err()
	if err != nil {
		return err
	}
	return nil
}

var _ him.SessionStorage = (*RedisStorage)(nil)

func KeySession(channel string) string {
	return fmt.Sprintf("login:sn:%s", channel)
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
