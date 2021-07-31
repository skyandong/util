package logger

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConf_Ensure(t *testing.T) {
	require := require.New(t)
	et := entry{
		FilePath:   "./log/trace",
		MaxSize:    512,
		MaxBackups: 1,
		MaxAge:     7,
	}
	c := make(Conf)
	c["logger"] = &et
	err := c.Ensure([]string{"logger"})
	require.NoError(err)
}

func TestConf_Get(t *testing.T) {
	et := entry{
		FilePath:   "./log/trace",
		MaxSize:    512,
		MaxBackups: 1,
		MaxAge:     7,
	}
	c := make(Conf)
	c["logger"] = &et
	lg := c.Get("logger")

	l := lg.GetLogger(InfoLevel)
	v1 := "haha"
	v2 := "xixi"
	v3 := "lala"
	l.Infow("test logger", "v1", v1, "v2", v2, "v3", v3)
	l.Warnw("test logger", "v1", v1, "v2", v2, "v3", v3)
	l.Errorw("test logger", "v1", v1, "v2", v2, "v3", v3)
}
