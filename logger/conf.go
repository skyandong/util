package logger

import (
	"fmt"
	"sync"
)

type entry struct {
	FilePath   string `yaml:"filePath"`   // 日志路径
	MaxSize    int    `yaml:"maxSize"`    // 单个日志最大的文件大小，单位: MB
	MaxBackups int    `yaml:"maxBackups"` // 日志文件最多保存多少个备份
	MaxAge     int    `yaml:"maxAge"`     // 文件最多保存多少天
	logger     *Logger
	init       sync.Once
}

// Conf for logger map
type Conf map[string]*entry

// Get logger by name
func (c Conf) Get(name string) (lg *Logger) {
	et, ok := c[name]
	if !ok {
		return
	}
	et.init.Do(func() {
		et.logger = New(InitConfig(et.FilePath, et.MaxSize, et.MaxBackups, et.MaxAge))
	})
	return et.logger
}

// Ensure loggers exist
func (c Conf) Ensure(names []string) error {
	for _, name := range names {
		if c.Get(name) == nil {
			return fmt.Errorf("%s logger not exist", name)
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
