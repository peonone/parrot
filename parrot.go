package parrot

import "github.com/go-redis/redis"
import "github.com/streadway/amqp"

//MakeRedisClient creates a redis client
func MakeRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

//MakeAMQPClient creates a AMQP channel
func MakeAMQPClient() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, nil, err
	}
	channel, err := conn.Channel()
	return conn, channel, err
}
