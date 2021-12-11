package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type entry struct {
	dsn    string
	client *mongo.Client
	init   sync.Once
}

// Conf for mongo map
type Conf map[string]*entry

// Get client by name
func (c Conf) Get(name string) (client *mongo.Client) {
	et, ok := c[name]
	if !ok {
		return
	}
	et.init.Do(func() {
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		et.client, err = mongo.Connect(ctx, options.Client().ApplyURI(et.dsn))
		if err != nil {
			//log.
		}
	})
	return et.client
}

// Ensure exist and reachable
func (c Conf) Ensure(names []string) error {
	for _, name := range names {
		client := c.Get(name)
		if client == nil {
			return fmt.Errorf("redis %s not exist", name)
		}
	}
	return nil
}

// Names returns keys in conf
func (c Conf) Names() []string {
	names := make([]string, 0, len(c))
	for key := range c {
		names = append(names, key)
	}
	return names
}
