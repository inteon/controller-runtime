/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package source

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	internal "sigs.k8s.io/controller-runtime/pkg/source/internal"

	"sigs.k8s.io/controller-runtime/pkg/cache"
)

// Source is a source of events (eh.g. Create, Update, Delete operations on Kubernetes Objects, Webhook callbacks, etc)
// which should be processed by event.EventHandlers to enqueue reconcile.Requests.
//
// * Use Kind for events originating in the cluster (e.g. Pod Create, Pod Update, Deployment Update).
//
// * Use Channel for events originating outside the cluster (eh.g. GitHub Webhook callback, Polling external urls).
//
// Users may build their own Source implementations.
type Source interface {
	// Start is internal and should be called only by the Controller to register an EventHandler with the Informer
	// to enqueue reconcile.Requests.
	Start(context.Context, workqueue.RateLimitingInterface) error
}

// SyncingSource is a source that needs syncing prior to being usable. The controller
// will call its WaitForSync prior to starting workers.
type SyncingSource interface {
	Source
	WaitForSync(ctx context.Context) error
}

// Kind creates a KindSource with the given cache provider.
func Kind(cache cache.Cache, object client.Object, eventhandler handler.EventHandler, prds ...predicate.Predicate) SyncingSource {
	return &internal.Kind{Cache: cache, Type: object, EventHandler: eventhandler, Predicates: prds}
}

// Informer creates an InformerSource with the given cache provider.
func Informer(informer cache.Informer, eventhandler handler.EventHandler, prds ...predicate.Predicate) Source {
	return &internal.Informer{Informer: informer, EventHandler: eventhandler, Predicates: prds}
}

// ChannelSource is a source of events originating outside the cluster (eh.g. GitHub Webhook callback).
// It can be used as input to the Channel function.
type ChannelSource = internal.Channel

// Channel creates a ChannelSource with the given buffer size.
func Channel(channel *ChannelSource, eventhandler handler.EventHandler, prds ...predicate.Predicate) Source {
	return &internal.ChannelHandle{Channel: channel, EventHandler: eventhandler, Predicates: prds}
}

// Func creates a FuncSource with the given function.
func Func(f internal.Func) Source {
	return f
}
