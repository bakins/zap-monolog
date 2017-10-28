package monolog

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func simpleProcessor() zapcore.Field {
	return zap.String("foo", "bar")
}

func TestNew(t *testing.T) {
	l, err := New()
	require.NotNil(t, l)
	require.Nil(t, err)

	l.Processors = append(l.Processors, simpleProcessor)
	l.Info("hello world")
}

func BenchmarkNew(b *testing.B) {
	l, _ := New()

	for n := 0; n < b.N; n++ {
		l.Info("hello world", zap.String("hello", "World"))
	}
}
