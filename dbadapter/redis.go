package dbadapter

import (
	"github.com/go-redis/redis"
)

const (
	keyAllGroup     = ":groups"
	keyGroupService = ":group:"
	keyServices     = ":services"
)

type RedisAdapter struct {
	client *redis.Client
	key    string
}

func NewRedisAdapter(config *Config, key string) DBAdapter {
	redisClient := redis.NewClient(&redis.Options{
		Addr:       config.Host,
		Password:   config.Password,
		DB:         0,
		MaxRetries: 3,
		OnConnect: func(conn *redis.Conn) error {
			return conn.Ping().Err()
		},
	})

	return &RedisAdapter{
		client: redisClient,
		key:    key,
	}
}

func (r *RedisAdapter) AddService(service string) error {
	return r.client.HSet(r.key + keyServices, service, "nil|nil").Err()
}

func (r *RedisAdapter) DeleteService(service string) error {
	return r.client.HDel(r.key + keyServices, service).Err()
}

func (r *RedisAdapter) DeployService(service string, branchAndModifier string) error {
	return r.client.HSet(r.key + keyServices, service, branchAndModifier).Err()
}

func (r *RedisAdapter) GetService(service string) (string, error) {
	res, err := r.client.HGet(r.key + keyServices, service).Result()
	if err != nil {
		return "", err
	}

	return res, nil
}

func (r *RedisAdapter) AddGroupService(group string, service string) error {
	return r.client.SAdd(r.key + keyGroupService + group, service).Err()
}

func (r *RedisAdapter) DeleteGroupService(group string, service string) error {
	return r.client.SRem(r.key + keyGroupService + group, service).Err()
}

func (r *RedisAdapter) GetAllServiceInGroup(group string) ([]string, error) {
	allService, err := r.client.SMembers(r.key + keyGroupService + group).Result()
	if err != nil {
		return nil, err
	}

	return allService, nil
}

func (r *RedisAdapter) AddGroup(group string) error {
	return r.client.SAdd(r.key + keyAllGroup, group).Err()
}

func (r *RedisAdapter) DeleteGroup(group string) error {
	return r.client.SRem(r.key + keyAllGroup, group).Err()
}

func (r *RedisAdapter) GetAllGroup() ([]string, error) {
	allGroup, err := r.client.SMembers(r.key + keyAllGroup).Result()
	if err != nil {
		return nil, err
	}

	return allGroup, nil
}

func (r *RedisAdapter) GetGroupServicesList(gn string) ([]string, error) {
	maps, err := r.client.HGetAll(r.key + gn).Result()
	if err != nil {
		return nil, err
	}

	var result []string
	for k := range maps {
		result = append(result, k)
	}

	return result, nil
}
