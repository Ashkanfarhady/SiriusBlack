package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	uuid "github.com/satori/go.uuid"
)

const MAX_TTL = 24 * time.Hour
const DEFAULT_COUNTER = 1

var ctx = context.Background()
var redisClient *redis.Client

type Server struct {
	BindAddress string
	Port        int
	RedisHost   string
	RedisPort   int
}

type Secret struct {
	Body    string `json:"body"`
	TTL     int    `json:"ttl"`
	Counter int    `json:"counter"`
}

type ResponseCreateSecret struct {
	Key string `json:"key"`
}

type ResponseGetSecret struct {
	Body string `json:"body"`
	TTL  int    `json:"ttl"`
}

func getCounter(key string) int {
	pipeline := redisClient.Pipeline()
	getCounter := pipeline.Get(ctx, fmt.Sprintf("counter_%s", key))
	pipeline.Decr(ctx, fmt.Sprintf("counter_%s", key))
	pipeline.Exec(ctx)
	counterStr, err := getCounter.Result()

	if err == redis.Nil {
		return -1
	}
	if err != nil {
		panic(err)
	}
	counter, err := strconv.Atoi(counterStr)
	if err != nil {
		panic(err)
	}
	return counter
}

func getSecret(key string) (string, time.Duration, bool) {
	secret, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		panic(err)
	}
	if secret == "" {
		return "", 0, false
	}
	ttl, err := redisClient.TTL(ctx, key).Result()
	if err != nil {
		panic(err)
	}
	return secret, ttl, true
}

func MethodAllowed(r *http.Request, method string) bool {
	if r.Method != method {
		return false
	}
	return true
}

func setSecret(key string, value string, ttl time.Duration, counter int) {
	if counter <= 0 {
		counter = DEFAULT_COUNTER
	}
	if ttl <= 0 {
		ttl = MAX_TTL
	}

	err := redisClient.Set(ctx, key, value, ttl).Err()
	if err != nil {
		panic(err)
	}

	err = redisClient.Set(ctx, fmt.Sprintf("counter_%s", key), counter, ttl).Err()
	if err != nil {
		panic(err)
	}
}

func GetSecretHandler(w http.ResponseWriter, r *http.Request) {
	if !MethodAllowed(r, http.MethodGet) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Path[1:] //Omitting starting slash

	if getCounter(key) <= 0 {
		w.WriteHeader(http.StatusNotFound)
		go func() {
			redisClient.Del(ctx, key, fmt.Sprintf("counter_%s", key))
		}()
		return
	}

	secret, ttl, exists := getSecret(key)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := ResponseGetSecret{Body: secret, TTL: int(ttl.Seconds())}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func CreateSecretHandler(w http.ResponseWriter, r *http.Request) {
	if !MethodAllowed(r, http.MethodPost) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var secret Secret
	err := json.NewDecoder(r.Body).Decode(&secret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	key := uuid.NewV4().String()
	setSecret(key, secret.Body, time.Duration(secret.TTL)*time.Minute, secret.Counter)

	response := ResponseCreateSecret{Key: key}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

}

func (s *Server) SetUpRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", s.RedisHost, s.RedisPort),
		Password: "",
		DB:       0,
	})

}

func (s *Server) Serve() {
	s.SetUpRedis()
	http.HandleFunc("/create/", CreateSecretHandler)
	http.HandleFunc("/", GetSecretHandler)
	http.ListenAndServe(fmt.Sprintf("%s:%d", s.BindAddress, s.Port), nil)
}
