package whitenoise

import (
	"context"
	"testing"
)

func TestDummy(t *testing.T) {
	ctx := context.Background()
	cfg := NewConfig()
	host, err := NewDummyHost(ctx, cfg)
	if err != nil {
		panic(err)
	}
	service, err := NewService(ctx, host, cfg)
	if err != nil {
		panic(err)
	}
	service.Start()
}
