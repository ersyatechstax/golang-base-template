package cache

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/redis.v5"

	redisclient "github.com/golang-base-template/util/cache/client"
)

const (
	CacheGBT = "redis-gbt"
)

type (
	// ResultType is a type for redis pipeline result type
	ResultType int

	//ICachePipeline is an interface for redis pipeline wrapper used by IMS repository
	ICachePipeline interface {
		NewPipeline(ctx context.Context, conn ...string) (txObj ICachePipeline, err error)
		MGet(ctx context.Context, keys ...string) (err error)
		MSet(ctx context.Context, pairs ...interface{}) (err error)
		HMGet(ctx context.Context, key string, field []string) (err error)
		HMSet(ctx context.Context, key string, data map[string]string) (err error)
		HDel(ctx context.Context, key string, field ...string) (err error)
		LRange(ctx context.Context, key string, start, stop int64) (err error)
		RPush(ctx context.Context, key string, data ...interface{}) (err error)
		SAdd(ctx context.Context, key string, data ...interface{}) (err error)
		SRem(ctx context.Context, key string, data ...interface{}) (err error)
		Del(ctx context.Context, key string) (err error)
		Expire(ctx context.Context, key string, expire time.Duration) (err error)
		Get(ctx context.Context, resType ResultType) (res interface{})
		Exec(ctx context.Context) (err error)
		HGetAll(ctx context.Context, keys string) (err error)
		Set(ctx context.Context, key string, value string, expire time.Duration) (err error)
		HSetNX(ctx context.Context, key, field, value string) (err error)
	}

	cachePipeline struct {
		conn     string
		rdsConn  *redis.Client
		pipe     *redis.Pipeline
		didExec  bool
		cmdCount int64

		// sliceCmds point to redis hash result
		sliceCmds []*redis.SliceCmd
		// stringSliceCmds point to redis list result
		stringSliceCmds []*redis.StringSliceCmd
		// statusCmds point to redis set command result
		statusCmds []*redis.StatusCmd
		// intCmds point to redis int result
		intCmds []*redis.IntCmd
		// boolCmds point to redis bool result
		boolCmds []*redis.BoolCmd
		// stringStringMapCmd point to get all result
		stringStringMapCmd []*redis.StringStringMapCmd
	}
)

// below is redis command definition supported on redis pipeline
var (
	pkgPipelineName = "cache.pipeline"

	// ResultHMGet is result type for HMGet
	ResultHMGet ResultType = 1
	// ResultHMSet is result type for HMSet
	ResultHMSet ResultType = 2
	// ResultLRANGE is result type for LRANGE
	ResultLRANGE ResultType = 3
	// ResultRPUSH is result type for RPUSH
	ResultRPUSH ResultType = 4
	// ResultDEL is result type for DEL
	ResultDEL ResultType = 5
	// ResultSADD is result type for SADD
	ResultSADD ResultType = 6
	// ResultMGet is result type for MGet
	ResultMGet ResultType = 7
	// ResultMSet is result type for MSet
	ResultMSet ResultType = 8
	// ResultEXPIRE is result type for Expire
	ResultEXPIRE ResultType = 9
	// ResultHGETALL is result type for HGETALL
	ResultHGETALL ResultType = 10
	// ResultHSetNX is result type for HSetNX
	ResultHSetNX ResultType = 11

	//ERROR VARIABLE

	// ErrFieldMissing is returned if some field is missing on redis hash cache
	ErrFieldMissing = errors.New("field is missing in redis cache")
	// ErrInvalidKeyorField is returned if input field or key is not specified
	ErrInvalidKeyorField = errors.New("no input field or key is specified")
	// ErrInvalidTypeResult is returned by formatter function (e.g. Hash, List, String, etc.) when formatting redis result
	ErrInvalidTypeResult = errors.New("invalid type result")
	// ErrConn is returned by NewPipeline function when initiate or test given connection is failed.
	ErrConn = errors.New("redis: cannot get connection to redis")
	// ErrUninitialized is not initialized. Please call NewPipeline() first before other command
	ErrUninitialized = errors.New("redis: pipeline is uninitialized")
)

// NewPkgPipeline return pipeline Obj interface
func NewPkgPipeline(conn ...string) (cp ICachePipeline) {
	cpObj := &cachePipeline{
		conn: CacheGBT,
	}

	if len(conn) == 0 {
		return cpObj
	}

	// assign manually
	cpObj.conn = conn[0]

	return cpObj
}

// NewPipeline starts a redis transaction.
// The function will check the connection to redis and it will return if error occured.
// To finish transaction, issue commit.
func (cp *cachePipeline) NewPipeline(ctx context.Context, conn ...string) (txObj ICachePipeline, err error) {

	connName := cp.conn
	if len(conn) > 0 {
		connName = conn[0]
	}

	// get connection
	rds, err := redisclient.GetConnection(connName)
	if err != nil {
		return nil, errors.Wrap(ErrConn, "[NewPipeline]")
	}

	return &cachePipeline{
		conn:            connName,
		rdsConn:         rds,
		pipe:            rds.Pipeline(),
		sliceCmds:       []*redis.SliceCmd{},
		stringSliceCmds: []*redis.StringSliceCmd{},
		statusCmds:      []*redis.StatusCmd{},
		intCmds:         []*redis.IntCmd{},
	}, nil
}

// MGet get some key with string type redis.
// Data will not be removed on redis, until commit is being issued.
func (cp *cachePipeline) MGet(ctx context.Context, keys ...string) (err error) {

	if cp.conn == "" {
		return errors.Wrap(ErrUninitialized, "[MGet]")
	}

	//skip wrong set
	if len(keys) == 0 {
		return errors.Wrap(ErrInvalidKeyorField, "[MGet]")
	}

	mget := cp.pipe.MGet(keys...)

	//set data
	cp.sliceCmds = append(cp.sliceCmds, mget)

	cp.cmdCount++
	return
}

// MSet set multiple key with string type redis.
// Data represent pair data (e.g. key value key value)
func (cp *cachePipeline) MSet(ctx context.Context, pairs ...interface{}) (err error) {

	if cp.conn == "" {
		return errors.Wrap(ErrUninitialized, "[MSet]")
	}

	//skip wrong set
	if len(pairs) == 0 {
		return errors.Wrap(ErrInvalidKeyorField, "[MSet]")
	}

	// unbalanced pair
	if len(pairs)%2 == 1 {
		return errors.Wrap(ErrInvalidKeyorField, "[MSet]")
	}

	mset := cp.pipe.MSet(pairs...)

	//set data
	cp.statusCmds = append(cp.statusCmds, mset)

	cp.cmdCount++
	return
}

// HMGet get some fields from key redis.
// Data will not be removed on redis, until commit is being issued.
func (cp *cachePipeline) HMGet(ctx context.Context, key string, fields []string) (err error) {

	if cp.conn == "" {
		return errors.Wrap(ErrUninitialized, "[HMGet]")
	}

	//skip wrong set
	if key == "" || len(fields) == 0 {
		return errors.Wrap(ErrInvalidKeyorField, "[HMGet]")
	}

	hmget := cp.pipe.HMGet(key, fields...)

	//set data
	cp.sliceCmds = append(cp.sliceCmds, hmget)

	cp.cmdCount++
	return
}

// HMSet set some fields from key redis.
func (cp *cachePipeline) HMSet(ctx context.Context, key string, data map[string]string) (err error) {

	//skip wrong set
	if key == "" || len(data) == 0 {
		return errors.Wrap(ErrInvalidKeyorField, "[HMSet]")
	}

	hmset := cp.pipe.HMSet(key, data)

	//set data

	cp.statusCmds = append(cp.statusCmds, hmset)
	cp.cmdCount++
	return
}

// HDel to delete field from key redis.
func (cp *cachePipeline) HDel(ctx context.Context, key string, fields ...string) (err error) {

	//skip wrong set
	if key == "" || len(fields) == 0 {
		return errors.Wrap(ErrInvalidKeyorField, "[HDel]")
	}

	hdel := cp.pipe.HDel(key, fields...)

	//set data
	cp.intCmds = append(cp.intCmds, hdel)
	cp.cmdCount++
	return
}

// LRange get list from key redis.
// To retrieve all data in list, use start 0 and stop -1.
func (cp *cachePipeline) LRange(ctx context.Context, key string, start, stop int64) (err error) {

	//skip wrong set
	if key == "" {
		return errors.Wrap(ErrInvalidKeyorField, "[LRange]")
	}

	lrange := cp.pipe.LRange(key, start, stop)

	//set data
	cp.stringSliceCmds = append(cp.stringSliceCmds, lrange)
	cp.cmdCount++
	return
}

// RPush set list to redis.
func (cp *cachePipeline) RPush(ctx context.Context, key string, data ...interface{}) (err error) {

	//skip wrong set
	if key == "" || len(data) == 0 {
		return errors.Wrap(ErrInvalidKeyorField, "[RPush]")
	}

	cpush := cp.pipe.RPush(key, data...)

	//set data
	cp.intCmds = append(cp.intCmds, cpush)
	cp.cmdCount++
	return
}

// SAdd add to set to redis.
func (cp *cachePipeline) SAdd(ctx context.Context, key string, data ...interface{}) (err error) {

	//skip wrong set
	if key == "" || len(data) == 0 {
		return errors.Wrap(ErrInvalidKeyorField, "[SAdd]")
	}

	sadd := cp.pipe.SAdd(key, data...)

	//set data
	cp.intCmds = append(cp.intCmds, sadd)
	cp.cmdCount++
	return
}

// SRem delete member data in specific key
func (cp *cachePipeline) SRem(ctx context.Context, key string, data ...interface{}) (err error) {

	//skip wrong set
	if key == "" || len(data) == 0 {
		return errors.Wrap(ErrInvalidKeyorField, "[SRem]")
	}

	srem := cp.pipe.SRem(key, data...)

	//set data
	cp.intCmds = append(cp.intCmds, srem)
	cp.cmdCount++
	return
}

// Del delete key from redis.
func (cp *cachePipeline) Del(ctx context.Context, key string) (err error) {

	//skip wrong set
	if key == "" {
		return errors.Wrap(ErrInvalidKeyorField, "[Del]")
	}

	del := cp.pipe.Del(key)

	//set data
	cp.intCmds = append(cp.intCmds, del)
	cp.cmdCount++
	return
}

// Expire set expire time to key redis.
func (cp *cachePipeline) Expire(ctx context.Context, key string, expire time.Duration) (err error) {

	//skip wrong set
	if key == "" {
		return errors.Wrap(ErrInvalidKeyorField, "[Expire]")
	}

	expRes := cp.pipe.Expire(key, expire)

	//set data
	cp.boolCmds = append(cp.boolCmds, expRes)
	cp.cmdCount++
	return
}

// Get finalize command that being pipelined
// If the exec has been running before, then it cannot be run anymore, please create a new pipeline instead
func (cp *cachePipeline) Get(ctx context.Context, resType ResultType) (res interface{}) {

	switch resType {
	case ResultHMGet:
		if len(cp.sliceCmds) > 0 {
			res, _ = cp.sliceCmds[0].Result()
			cp.sliceCmds = cp.sliceCmds[1:]
		}
	case ResultHMSet:
		if len(cp.statusCmds) > 0 {
			res, _ = cp.statusCmds[0].Result()
			cp.statusCmds = cp.statusCmds[1:]
		}
	case ResultLRANGE:
		if len(cp.stringSliceCmds) > 0 {
			res, _ = cp.stringSliceCmds[0].Result()
			cp.stringSliceCmds = cp.stringSliceCmds[1:]
		}
	case ResultRPUSH:
		if len(cp.intCmds) > 0 {
			res, _ = cp.intCmds[0].Result()
			cp.intCmds = cp.intCmds[1:]
		}
	case ResultDEL:
		if len(cp.intCmds) > 0 {
			res, _ = cp.intCmds[0].Result()
			cp.intCmds = cp.intCmds[1:]
		}
	case ResultSADD:
		if len(cp.intCmds) > 0 {
			res, _ = cp.intCmds[0].Result()
			cp.intCmds = cp.intCmds[1:]
		}
	case ResultMGet:
		if len(cp.sliceCmds) > 0 {
			res, _ = cp.sliceCmds[0].Result()
			cp.sliceCmds = cp.sliceCmds[1:]
		}
	case ResultMSet:
		if len(cp.statusCmds) > 0 {
			res, _ = cp.statusCmds[0].Result()
			cp.statusCmds = cp.statusCmds[1:]
		}
	case ResultEXPIRE:
		if len(cp.boolCmds) > 0 {
			res, _ = cp.boolCmds[0].Result()
			cp.boolCmds = cp.boolCmds[1:]
		}
	case ResultHGETALL:
		if len(cp.stringStringMapCmd) > 0 {
			res, _ = cp.stringStringMapCmd[0].Result()
			cp.stringStringMapCmd = cp.stringStringMapCmd[1:]
		}
	case ResultHSetNX:
		if len(cp.boolCmds) > 0 {
			res, _ = cp.boolCmds[0].Result()
			cp.boolCmds = cp.boolCmds[1:]
		}
	default:
		log.Println("[RedisPipeline][Get]result type is not defined", resType)
	}

	return
}

// Exec finalize command that being pipelined
// If the exec has been running before, then it cannot be run anymore, please create a new pipeline instead
func (cp *cachePipeline) Exec(ctx context.Context) (err error) {
	if cp.didExec == true {
		return nil
	}
	if cp.cmdCount == 0 {
		return nil
	}

	_, err = cp.pipe.Exec()
	if err != nil {
		return errors.Wrapf(err, "[RedisPipeline][Exec] error when execute redis command in pipeline")
	}
	cp.didExec = true

	return
}

// HGetAll get all fields from redis hash
func (cp *cachePipeline) HGetAll(ctx context.Context, key string) (err error) {

	if cp.conn == "" {
		return errors.Wrap(ErrUninitialized, "[HGetAll]")
	}

	//skip wrong set
	if key == "" {
		return errors.Wrap(ErrInvalidKeyorField, "[HGetAll]")
	}

	hgetall := cp.pipe.HGetAll(key)

	//set data
	cp.stringStringMapCmd = append(cp.stringStringMapCmd, hgetall)

	cp.cmdCount++
	return
}

// Set will set key value to redis
func (cp *cachePipeline) Set(ctx context.Context, key string, value string, expire time.Duration) (err error) {

	//skip wrong set
	if key == "" || value == "" {
		return errors.Wrap(ErrInvalidKeyorField, "[Set]")
	}

	sadd := cp.pipe.Set(key, value, expire)

	//set data
	cp.statusCmds = append(cp.statusCmds, sadd)
	cp.cmdCount++
	return
}

// HSetNX will set key value to redis if key not exists
func (cp *cachePipeline) HSetNX(ctx context.Context, key, field, value string) (err error) {

	//skip wrong set
	if key == "" || field == "" {
		return errors.Wrap(ErrInvalidKeyorField, "[HSetNX]")
	}

	hSetNX := cp.pipe.HSetNX(key, field, value)

	cp.boolCmds = append(cp.boolCmds, hSetNX)
	cp.cmdCount++
	return
}

// Hash return result with hash format (map[string]string)
func Hash(source interface{}, fields []string) (result map[string]string, err error) {
	result = make(map[string]string, 0)
	if source == nil {
		return result, ErrInvalidTypeResult
	}
	rt := reflect.TypeOf(source)
	if rt.Kind() != reflect.Slice ||
		(rt.Kind() == reflect.Slice && rt.Elem().Kind() != reflect.Interface) {
		return result, errors.Wrapf(ErrInvalidTypeResult, "[RedisPipeline][Hash] Type is not a slice of interface : %v", rt.Name())
	}
	values := source.([]interface{})

	missing := []string{}
	for i := range fields {
		if i >= len(values) || values[i] == nil {
			missing = append(missing, fields[i])
			continue
		}
		result[fields[i]] = values[i].(string)
	}

	if len(missing) > 0 {
		err = errors.Wrapf(ErrFieldMissing, "[RedisPipeline][Hash] missing field : %v", missing)
		return result, err
	}
	return result, err
}

// List return result with list string format ([]string)
func List(source interface{}) (result []string, err error) {
	result = make([]string, 0)
	if source == nil {
		return result, ErrInvalidTypeResult
	}
	rt := reflect.TypeOf(source)
	if rt.Kind() != reflect.Slice {
		return result, errors.Wrapf(ErrInvalidTypeResult, "[RedisPipeline][List] Type is not slice : %v", rt.Name())
	}

	switch rt.Elem().Kind() {

	case reflect.String: // list string
		result = source.([]string)
	case reflect.Interface: // list interafce
		resInterface := source.([]interface{})
		for _, res := range resInterface {
			result = append(result, fmt.Sprintf("%v", res))
		}

	}

	return
}

// String return result with list string format ([]string)
func String(source interface{}) (result string, err error) {
	if source == nil {
		return result, ErrInvalidTypeResult
	}
	rt := reflect.TypeOf(source)
	if rt.Kind() != reflect.String {
		return result, errors.Wrapf(ErrInvalidTypeResult, "[RedisPipeline][String] Type is not string : %v", rt.Name())
	}
	result = source.(string)
	return
}

// Int return result with int format ([]string)
func Int(source interface{}) (result int64, err error) {
	if source == nil {
		return result, ErrInvalidTypeResult
	}
	rt := reflect.TypeOf(source)
	if rt.Kind() != reflect.String {
		return result, errors.Wrapf(ErrInvalidTypeResult, "[RedisPipeline][Int] Type is not string : %v", rt.Name())
	}
	result, err = strconv.ParseInt(source.(string), 10, 64)
	return
}

// Bool return result with bool format ([]string)
func Bool(source interface{}) (result bool, err error) {
	if source == nil {
		return result, ErrInvalidTypeResult
	}
	result, ok := source.(bool)
	if !ok {
		return result, errors.Wrapf(ErrInvalidTypeResult, "[RedisPipeline][Bool] Type is not bool : %v", source)
	}
	return
}

// HashMap return result with Hash format, convert HGETALL result
func HashMap(source interface{}, fields []string, optFields ...[]string) (result map[string]string, err error) {
	result = make(map[string]string, 0)
	if source == nil {
		return result, ErrInvalidTypeResult
	}
	rt := reflect.TypeOf(source)
	if rt.Kind() != reflect.Map ||
		(rt.Kind() == reflect.Map && rt.Elem().Kind() != reflect.String) {
		return result, errors.Wrapf(ErrInvalidTypeResult, "[RedisPipeline][HashMap] Type is not map of string : %v", rt.Name())
	}
	values := source.(map[string]string)

	// return all values if no fields given
	if len(fields) == 0 {
		return values, nil
	}

	missing := []string{}
	for i, field := range fields {
		if i >= len(values) || values[field] == "" {
			missing = append(missing, fields[i])
			continue
		}
		result[fields[i]] = values[field]
	}

	// add support for optional field retrieval
	var optionalField []string
	if len(optFields) > 0 {
		optionalField = optFields[0]
	}
	for i, field := range optionalField {
		if i >= len(values) || values[field] == "" {
			continue
		}
		result[field] = values[field]
	}

	if len(missing) > 0 {
		err = errors.Wrapf(ErrFieldMissing, "[RedisPipeline][HashMap] missing field : %v", missing)
		return result, err
	}
	return result, err
}
