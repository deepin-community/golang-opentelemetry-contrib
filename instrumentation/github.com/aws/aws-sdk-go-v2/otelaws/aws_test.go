// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otelaws

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/otel/propagation"
)

type mockPropagator struct {
	injectKey   string
	injectValue string
}

func (p mockPropagator) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	carrier.Set(p.injectKey, p.injectValue)
}

func (p mockPropagator) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return context.TODO()
}

func (p mockPropagator) Fields() []string {
	return []string{}
}

func Test_otelMiddlewares_finalizeMiddleware(t *testing.T) {
	stack := middleware.Stack{
		Finalize: middleware.NewFinalizeStep(),
	}

	propagator := mockPropagator{
		injectKey:   "mock-key",
		injectValue: "mock-value",
	}

	m := otelMiddlewares{
		propagator: propagator,
	}

	err := m.finalizeMiddleware(&stack)
	require.NoError(t, err)

	input := &smithyhttp.Request{
		Request: &http.Request{
			Header: http.Header{},
		},
	}

	next := middleware.HandlerFunc(func(ctx context.Context, input interface{}) (output interface{}, metadata middleware.Metadata, err error) {
		return nil, middleware.Metadata{}, nil
	})

	_, _, _ = stack.Finalize.HandleMiddleware(context.Background(), input, next)

	// Assert header has been updated with injected values
	key := http.CanonicalHeaderKey(propagator.injectKey)
	value := propagator.injectValue

	assert.Contains(t, input.Header, key)
	assert.Contains(t, input.Header[key], value)
}

func Test_Span_name(t *testing.T) {
	serviceID1 := ""
	serviceID2 := "ServiceID"
	operation1 := ""
	operation2 := "Operation"

	assert.Equal(t, "", spanName(serviceID1, operation1))
	assert.Equal(t, spanName(serviceID1, operation2), "."+operation2)
	assert.Equal(t, spanName(serviceID2, operation1), serviceID2)
	assert.Equal(t, spanName(serviceID2, operation2), serviceID2+"."+operation2)
}
