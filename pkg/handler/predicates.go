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

package handler

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// WithPredicates returns an EventHandler that only calls the underlying handler if all the given predicates return true.
func WithPredicates[T any](handler TypedEventHandler[T], predicates ...predicate.TypedPredicate[T]) TypedEventHandler[T] {
	return &TypedFuncs[T]{
		CreateFunc: func(ctx context.Context, createEvent event.TypedCreateEvent[T], queue workqueue.RateLimitingInterface) {
			for _, predicate := range predicates {
				if !predicate.Create(createEvent) {
					return
				}
			}
			handler.Create(ctx, createEvent, queue)
		},
		UpdateFunc: func(ctx context.Context, updateEvent event.TypedUpdateEvent[T], queue workqueue.RateLimitingInterface) {
			for _, predicate := range predicates {
				if !predicate.Update(updateEvent) {
					return
				}
			}
			handler.Update(ctx, updateEvent, queue)
		},
		DeleteFunc: func(ctx context.Context, deleteEvent event.TypedDeleteEvent[T], queue workqueue.RateLimitingInterface) {
			for _, predicate := range predicates {
				if !predicate.Delete(deleteEvent) {
					return
				}
			}
			handler.Delete(ctx, deleteEvent, queue)
		},
		GenericFunc: func(ctx context.Context, genericEvent event.TypedGenericEvent[T], queue workqueue.RateLimitingInterface) {
			for _, predicate := range predicates {
				if !predicate.Generic(genericEvent) {
					return
				}
			}
			handler.Generic(ctx, genericEvent, queue)
		},
	}
}
