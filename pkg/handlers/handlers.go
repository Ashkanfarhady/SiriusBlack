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
}

func GetSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Path[1:]
	pipeline := redisClient.Pipeline()
	getCounter := pipeline.Get(ctx, fmt.Sprintf("counter_%s", key))
	pipeline.Decr(ctx, fmt.Sprintf("counter_%s", key))
	pipeline.Exec(ctx)
	counterStr, err := getCounter.Result()
	counter, err := strconv.Atoi(counterStr)
	if err != nil {
		panic(err)
	}
	if counter <= 0 {
		w.WriteHeader(http.StatusNotFound)
		go func() {
			redisClient.Del(ctx, key, fmt.Sprintf("counter_%s", key))
		}()
		return
	}
	secret, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		panic(err)
	}
	if secret == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := ResponseGetSecret{Body: secret}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func CreateSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var secret Secret
	err := json.NewDecoder(r.Body).Decode(&secret)

	if err != nil {
		w.WriteHeader(400)
	}
	key := uuid.NewV4().String()
	err = redisClient.Set(ctx, key, secret.Body, time.Duration(secret.TTL)*time.Minute).Err()
	if err != nil {
		panic(err)
	}
	err = redisClient.Set(ctx, fmt.Sprintf("counter_%s", key), secret.Counter, time.Duration(secret.TTL)*time.Minute).Err()
	if err != nil {
		panic(err)
	}
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
