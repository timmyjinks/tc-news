package store

import (
	"context"
	"fmt"
)

func (s *RedisStore) Exists(key string) (bool, error) {
	count, err := s.db.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (s *RedisStore) Get(key string) (map[string]string, error) {
	return s.db.HGetAll(context.Background(), key).Result()
}

func (s *RedisStore) Create(key string, value Data) error {
	res := s.db.HSet(context.Background(), key, map[string]interface{}{
		"id":   value.Id,
		"name": value.Name,
	})
	s.db.Expire(context.Background(), key, value.TTL)
	fmt.Println(res.Result())
	return res.Err()
}

func (s *RedisStore) Delete(key string) error {
	res := s.db.Del(context.Background(), key)
	return res.Err()
}
