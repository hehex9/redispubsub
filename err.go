package redispubsub

import "github.com/redis/go-redis/v9"

func errorAs(err error, i interface{}) bool {
	switch terr := err.(type) {
	case redis.Error:
		if p, ok := i.(*redis.Error); ok {
			*p = terr
			return true
		}
	}
	return false
}
