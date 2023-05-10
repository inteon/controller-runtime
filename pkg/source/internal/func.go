package internal

import (
	"context"
	"fmt"

	"k8s.io/client-go/util/workqueue"
)

// Func is a function that implements Source.
type Func func(context.Context, workqueue.RateLimitingInterface) error

// Start implements Source.
func (f Func) Start(ctx context.Context, queue workqueue.RateLimitingInterface) error {
	return f(ctx, queue)
}

func (f Func) String() string {
	return fmt.Sprintf("func source: %p", f)
}
