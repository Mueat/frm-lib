package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gitee.com/Rainkropy/frm-lib/log"
	"gitee.com/Rainkropy/frm-lib/util"
	"github.com/go-redis/redis/v8"
)

var (
	ctx   = context.Background()
	pools map[string]Pools
)

type Pools struct {
	Prefix string
	Config RedisConfig
	client *redis.Client
}

type RedisConfig struct {
	Network      string //连接类型 tcp or unix
	Addr         string //地址
	Username     string //用户名，redis6.0以上
	Password     string //密码
	Prefix       string //key的前缀
	DB           int64  //数据库
	PoolSize     int64  //最大链接数
	MinIdleConns int64  //保持的最小链接数
	IdleTimeout  int64  //链接过期时间，单位：秒
	Default      bool   //是否是默认的redis
}

// 初始化
func InitRedis(configs map[string]RedisConfig) {
	pools = make(map[string]Pools)
	for k, conf := range configs {
		opts := &redis.Options{
			Network:      conf.Network,
			Addr:         conf.Addr,
			MinIdleConns: int(conf.MinIdleConns),
			IdleTimeout:  time.Duration(conf.IdleTimeout) * time.Second,
			PoolSize:     int(conf.PoolSize),
			DB:           int(conf.DB),
		}
		if conf.Username != "" {
			opts.Username = conf.Username
		}
		if conf.Password != "" {
			opts.Password = conf.Password
		}

		pool := Pools{
			client: redis.NewClient(opts),
			Config: conf,
			Prefix: conf.Prefix,
		}
		pools[k] = pool
	}
}

// 获取redis链接
func GetRedis(name string) *Pools {
	if name != "" {
		if pool, ok := pools[name]; ok {
			return &pool
		}
	}

	for _, pool := range pools {
		if pool.Config.Default {
			return &pool
		}
	}

	return nil
}

// 获取默认连接
func GetDefaultRedis() *Pools {
	return GetRedis("")
}

func (r *Pools) GetKey(key string) string {
	return fmt.Sprintf("%s%s", r.Prefix, key)
}

func (r *Pools) GetClient() *redis.Client {
	return r.client
}

func (r *Pools) Set(k, v string, ex time.Duration) error {
	k = r.GetKey(k)
	err := r.client.Set(ctx, k, v, ex).Err()
	if err != nil {
		log.Error().Err(err).Msgf("redis set error key: %s value : %s  error:%s", k, v, err)
		return err
	}
	return nil
}

func (r *Pools) Del(k string) error {
	k = r.GetKey(k)
	err := r.client.Del(ctx, k).Err()
	if err != nil {
		log.Error().Err(err).Msgf("redis Del error key: %v  error:%s", k, err)
		return err
	}
	return nil
}

func (r *Pools) SetNXEX(k, v string, ex time.Duration) (bool, error) {
	k = r.GetKey(k)
	res, err := r.client.SetNX(ctx, k, v, ex).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis SetNXEX error key: %s value : %s  error:%s", k, v, err)
		return false, err
	}
	return res, nil
}
func (r *Pools) Expire(k string, ex time.Duration) error {
	k = r.GetKey(k)
	err := r.client.Expire(ctx, k, ex).Err()
	if err != nil {
		log.Error().Err(err).Msgf("redis Expire error key: %s ex : %s  error:%s", k, ex, err)
		return err
	}
	return nil
}
func (r *Pools) GetString(k string) string {
	k = r.GetKey(k)
	res, err := r.client.Get(ctx, k).Result()

	if err != nil {
		log.Error().Err(err).Msgf("redis get error key: %s  error:%s", k, err.Error())
	}
	return res
}

func (r *Pools) BatchPushQueue(k string, values []string) (err error) {
	if len(values) == 0 {
		return
	}
	k = r.GetKey(k)
	err = r.client.LPush(ctx, k, values).Err()
	if err != nil {
		log.Error().Err(err).Msgf("redis LPUSH key: %s value: %v error: %s", k, values, err)
	}
	return
}

func (r *Pools) PopQueue(k string, timeout time.Duration) (data string, err error) {
	k = r.GetKey(k)
	nameAndData, err := r.client.BRPop(ctx, timeout, k).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis BRPOP queue queueName %s error %v ", k, err.Error())
		return "", err
	}
	if len(nameAndData) > 1 {
		data = nameAndData[1]
	}
	return data, nil
}

func (r *Pools) LPush(k string, v string) error {
	k = r.GetKey(k)
	err := r.client.LPush(ctx, k, v).Err()
	if err != nil {
		log.Error().Err(err).Msgf("redis LPUSH key : %s value : %s error : %s", k, v, err.Error())
		return err
	}
	return nil
}

func (r *Pools) HGet(key, field string) (string, error) {
	key = r.GetKey(key)
	res, err := r.client.HGet(ctx, key, field).Result()
	if err != nil && err != redis.Nil {
		log.Error().Err(err).Msgf("redis HGET key : %s field : %s error : %s", key, field, err.Error())
		return "", err
	}
	return res, nil
}

func (r *Pools) HGetAll(k string) (map[string]string, error) {
	k = r.GetKey(k)
	res, err := r.client.HGetAll(ctx, k).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis HGetAll key : %s  error : %v", k, err.Error())
		return nil, err
	}
	return res, nil
}

func (r *Pools) SMembers(key string) ([]string, error) {
	key = r.GetKey(key)
	res, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis HGetAll key : %s error : %v ", key, err.Error())
		return nil, err
	}
	return res, nil
}

//hlen
func (r *Pools) HLen(key string) (int64, error) {
	key = r.GetKey(key)
	res, err := r.client.HLen(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis HLen key : %s  error : %s", key, err.Error())
		return res, err
	}
	return res, nil
}
func (r *Pools) HSet(key, field string, value string) error {
	key = r.GetKey(key)
	err := r.client.HSet(ctx, key, field, value).Err()
	if err != nil {
		log.Error().Err(err).Msgf("redis HSET key : %s  field : %s  value : %s error : %s", key, field, value, err.Error())
		return err
	}
	return nil
}

// HMSet command:
// Sets the specified fields to their respective values in the hash stored at key.
// This command overwrites any existing fields in the hash.
// If key does not exist, a new key holding a hash is created.
func (r *Pools) HMSet(key string, values map[string]interface{}) error {
	key = r.GetKey(key)
	err := r.client.HMSet(ctx, key, values).Err()
	if err != nil {
		log.Error().Err(err).Msgf("redis HMSET key : %s   value : %v error : %s", key, values, err.Error())
		return err
	}
	return nil
}

// HDel command:
func (r *Pools) HDel(key string, fields []string) error {
	key = r.GetKey(key)
	err := r.client.HDel(ctx, key, fields...).Err()
	if err != nil {
		log.Error().Err(err).Msgf("redis HDEL key : %s  fields : %v  error : %s", key, fields, err.Error())
		return err
	}
	return nil
}

// zdd command:
func (r *Pools) ZAdd(key string, score int64, member interface{}) error {
	key = r.GetKey(key)
	err := r.client.ZAdd(ctx, key, &redis.Z{Score: float64(score), Member: member}).Err()
	if err != nil {
		log.Error().Err(err).Msgf("redis ZAdd key : %s  score : %v  member : %v error : %s", key, score, member, err.Error())
		return err
	}
	return nil
}

// zdd command:
func (r *Pools) Exists(key string) (bool, error) {
	key = r.GetKey(key)
	res, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis Exists key : %s  error : %s", key, err.Error())
		return false, err
	}
	if res > 0 {
		return true, nil
	}
	return false, nil
}

//获取整个集合元素
func (r *Pools) ZRangeAll(key string) ([]string, error) {
	key = r.GetKey(key)
	res, err := r.client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis Exists key : %s  error : %s", key, err.Error())
		return res, err
	}
	return res, nil
}

//删除集合元素
func (r *Pools) ZRem(key string, members []string) error {
	key = r.GetKey(key)
	err := r.client.ZRem(ctx, key, members).Err()
	if err != nil {
		log.Error().Err(err).Msgf("redis Exists key : %s  error : %s", key, err.Error())
		return err
	}
	return nil
}

//删除集合元素
func (r *Pools) ZCard(key string) (int64, error) {
	key = r.GetKey(key)
	res, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis ZCard key : %s  error : %s", key, err.Error())
		return res, err
	}
	return res, nil
}

func (r *Pools) Incr(key string) (int64, error) {
	key = r.GetKey(key)
	res, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis ZCard key : %s  error : %s", key, err.Error())
		return res, err
	}
	return res, nil
}

func (r *Pools) Decr(key string) (int64, error) {
	key = r.GetKey(key)
	res, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis ZCard key : %s  error : %s", key, err.Error())
		return res, err
	}
	return res, nil
}

// bitmap
func (r *Pools) SetBit(key string, offset int64, val int) (int64, error) {
	key = r.GetKey(key)
	res, err := r.client.SetBit(ctx, key, offset, val).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis SetBit key : %s  error : %s", key, err.Error())
		return res, err
	}
	return res, nil
}

func (r *Pools) GetBit(key string, offset int64) (int64, error) {
	key = r.GetKey(key)
	res, err := r.client.GetBit(ctx, key, offset).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis GetBit key : %s  error : %s", key, err.Error())
		return res, err
	}
	return res, nil
}

func (r *Pools) BitCount(key string, start int64, end int64) (int64, error) {
	key = r.GetKey(key)
	bc := redis.BitCount{
		Start: start,
		End:   end,
	}
	res, err := r.client.BitCount(ctx, key, &bc).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis BitCount key : %s  error : %s", key, err.Error())
		return res, err
	}
	return res, nil
}

func (r *Pools) BitOpAnd(destKey string, keys ...string) (int64, error) {
	destKey = r.GetKey(destKey)
	mkeys := make([]string, 0)
	for _, ky := range keys {
		mkeys = append(mkeys, r.GetKey(ky))
	}
	res, err := r.client.BitOpAnd(ctx, destKey, mkeys...).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis BitOpAnd keys : %v  error : %s", keys, err.Error())
		return res, err
	}
	return res, nil
}

func (r *Pools) BitOpOr(destKey string, keys ...string) (int64, error) {
	destKey = r.GetKey(destKey)
	mkeys := make([]string, 0)
	for _, ky := range keys {
		mkeys = append(mkeys, r.GetKey(ky))
	}
	res, err := r.client.BitOpOr(ctx, destKey, mkeys...).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis BitOpOr keys : %v  error : %s", keys, err.Error())
		return res, err
	}
	return res, nil
}

func (r *Pools) BitOpXor(destKey string, keys ...string) (int64, error) {
	destKey = r.GetKey(destKey)
	mkeys := make([]string, 0)
	for _, ky := range keys {
		mkeys = append(mkeys, r.GetKey(ky))
	}
	res, err := r.client.BitOpXor(ctx, destKey, mkeys...).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis BitOpXor keys : %v  error : %s", keys, err.Error())
		return res, err
	}
	return res, nil
}

func (r *Pools) BitOpNot(destKey string, key string) (int64, error) {
	destKey = r.GetKey(destKey)
	key = r.GetKey(key)
	res, err := r.client.BitOpNot(ctx, destKey, key).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis BitOpNot destkey: %s key : %s  error : %s", destKey, key, err.Error())
		return res, err
	}
	return res, nil
}

func (r *Pools) BitPos(key string, bit int64, pos ...int64) (int64, error) {
	key = r.GetKey(key)
	res, err := r.client.BitPos(ctx, key, bit, pos...).Result()
	if err != nil {
		log.Error().Err(err).Msgf("redis BitPos key : %s  error : %s", key, err.Error())
		return res, err
	}
	return res, nil
}

// Pipeline
func (r *Pools) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// QueryCache
func (r *Pools) Query(key string, data interface{}, fields ...interface{}) error {
	formats := make([]string, 0)
	for range fields {
		formats = append(formats, "%v")
	}
	format := util.Base64Encode(util.Implode("_", formats))
	field := fmt.Sprintf(format, fields...)
	cacheStr, err := r.HGet(key, field)
	if err != nil {
		return err
	}
	if cacheStr != "" {
		err := json.Unmarshal([]byte(cacheStr), data)
		if err == nil {
			return err
		}
	} else {
		return errors.New("CacheNotFound")
	}
	return nil
}

func (r *Pools) SaveQuery(key string, data interface{}, fields ...interface{}) error {
	formats := make([]string, 0)
	for range fields {
		formats = append(formats, "%v")
	}
	format := util.Base64Encode(util.Implode("_", formats))
	field := fmt.Sprintf(format, fields...)
	cacheStr, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return r.HSet(key, field, string(cacheStr))
}
