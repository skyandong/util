package mysql

import (
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type entry struct {
	Driver string
	Dsn    string
	DB     *gorm.DB
	init   sync.Once
}

// Drivers
const (
	DriverMysql  = "mysql"
	DriverSqlite = "sqlite"

	defaultDriver = DriverMysql
)

// Conf for mongo map
type Conf map[string]*entry

// Get client by name
func (c Conf)Get(name string) (db *gorm.DB) {
	et, ok := c[name]
	if !ok {
		return
	}
	et.init.Do(func() {
		var err error
		et.DB, err = gorm.Open(et.Driver, et.Dsn)
		if err != nil {
			//log.
		}
		if err = et.DB.DB().Ping();err!=nil{
			_ = et.DB.Close()
			return
		}
	})
	return et.DB
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