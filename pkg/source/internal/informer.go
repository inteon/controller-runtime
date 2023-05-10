package internal

import (
	"context"
	"fmt"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Informer is used to provide a source of events originating inside the cluster from Watches (e.g. Pod Create).
type Informer struct {
	// Informer is the controller-runtime Informer
	Informer cache.Informer

	// EventHandler is the handler to call when events are received.
	EventHandler handler.EventHandler

	// Predicates are the predicates to evaluate whether to handle the event.
	Predicates []predicate.Predicate
}

// Start is internal and should be called only by the Controller to register an EventHandler with the Informer
// to enqueue reconcile.Requests.
func (is *Informer) Start(ctx context.Context, queue workqueue.RateLimitingInterface) error {
	if is.Informer == nil {
		return fmt.Errorf("must create Informer with a non-nil Informer")
	}
	if is.EventHandler == nil {
		return fmt.Errorf("must create Informer with a non-nil EventHandler")
	}

	_, err := is.Informer.AddEventHandler(NewEventHandler(ctx, queue, is.EventHandler, is.Predicates).HandlerFuncs())
	if err != nil {
		return err
	}
	return nil
}

func (is *Informer) String() string {
	return fmt.Sprintf("informer source: %p", is.Informer)
}
