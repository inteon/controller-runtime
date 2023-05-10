package internal

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	// defaultBufferSize is the default number of event notifications that can be buffered.
	defaultBufferSize = 1024
)

// Channel is used to provide a source of events originating outside the cluster
// (e.g. GitHub Webhook callback).  Channel requires the user to wire the external
// source (eh.g. http handler) to write GenericEvents to the underlying channel.
type Channel struct {
	// once ensures the event distribution goroutine will be performed only once
	once sync.Once

	// Source is the source channel to fetch GenericEvents
	Source <-chan event.GenericEvent

	// DestBufferSize is the buffer size of the destination channels
	DestBufferSize int

	// dest is the destination channels of the added event handlers
	dest []chan event.GenericEvent

	// destLock is to ensure the destination channels are safely added/removed
	destLock sync.Mutex
}

func (cs *Channel) start(ctx context.Context) error {
	if cs.Source == nil {
		return fmt.Errorf("must create Channel with a non-nil Source")
	}

	cs.once.Do(func() {
		// Distribute GenericEvents to all EventHandler / Queue pairs Watching this source
		go cs.syncLoop(ctx)
	})

	return nil
}

func (cs *Channel) doStop() {
	cs.destLock.Lock()
	defer cs.destLock.Unlock()

	for _, dst := range cs.dest {
		close(dst)
	}
}

func (cs *Channel) distribute(evt event.GenericEvent) {
	cs.destLock.Lock()
	defer cs.destLock.Unlock()

	for _, dst := range cs.dest {
		// We cannot make it under goroutine here, or we'll meet the
		// race condition of writing message to closed channels.
		// To avoid blocking, the dest channels are expected to be of
		// proper buffer size. If we still see it blocked, then
		// the controller is thought to be in an abnormal state.
		dst <- evt
	}
}

func (cs *Channel) syncLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Close destination channels
			cs.doStop()
			return
		case evt, stillOpen := <-cs.Source:
			if !stillOpen {
				// if the source channel is closed, we're never gonna get
				// anything more on it, so stop & bail
				cs.doStop()
				return
			}
			cs.distribute(evt)
		}
	}
}

func (cs *Channel) addHandle() <-chan event.GenericEvent {
	cs.destLock.Lock()
	defer cs.destLock.Unlock()

	if cs.DestBufferSize <= 0 {
		cs.DestBufferSize = defaultBufferSize
	}

	dst := make(chan event.GenericEvent, cs.DestBufferSize)
	cs.dest = append(cs.dest, dst)

	return dst
}

// ChannelHandle is a handle to a Channel. It can be used as a Source for a
// Controller.
type ChannelHandle struct {
	Channel *Channel

	// EventHandler is the handler to call when events are received.
	EventHandler handler.EventHandler

	// Predicates are the predicates to evaluate whether to handle the event.
	Predicates []predicate.Predicate

	// DestBufferSize is the specified buffer size of the ChannelHandle.
	// Default to 1024 if not specified.
	DestBufferSize int
}

func (cs *ChannelHandle) String() string {
	return fmt.Sprintf("channel source: %p", cs)
}

// Start implements Source and should only be called by the Controller.
func (cs *ChannelHandle) Start(ctx context.Context, queue workqueue.RateLimitingInterface) error {
	if cs.EventHandler == nil {
		return fmt.Errorf("must create ChannelHandle with a non-nil EventHandler")
	}
	if cs.Channel == nil {
		return fmt.Errorf("must create ChannelHandle with a non-nil Channel")
	}

	if err := cs.Channel.start(ctx); err != nil {
		return err
	}

	dst := cs.Channel.addHandle()

	go func() {
		for evt := range dst {
			shouldHandle := true
			for _, p := range cs.Predicates {
				if !p.Generic(evt) {
					shouldHandle = false
					break
				}
			}

			if shouldHandle {
				func() {
					ctx, cancel := context.WithCancel(ctx)
					defer cancel()
					cs.EventHandler.Generic(ctx, evt, queue)
				}()
			}
		}
	}()

	return nil
}
