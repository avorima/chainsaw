package runner

import (
	"context"

	"github.com/kyverno/chainsaw/pkg/apis"
	"github.com/kyverno/chainsaw/pkg/apis/v1alpha1"
	"github.com/kyverno/chainsaw/pkg/apis/v1alpha2"
	"github.com/kyverno/chainsaw/pkg/client"
	fake "github.com/kyverno/chainsaw/pkg/client/testing"
	"github.com/kyverno/chainsaw/pkg/discovery"
	"github.com/kyverno/chainsaw/pkg/model"
	enginecontext "github.com/kyverno/chainsaw/pkg/runner/context"
	"github.com/kyverno/chainsaw/pkg/runner/mocks"
	"github.com/kyverno/chainsaw/pkg/testing"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/clock"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestTestsProcessor_Run(t *testing.T) {
	testCases := []struct {
		name         string
		config       model.Configuration
		client       client.Client
		clock        clock.PassiveClock
		bindings     apis.Bindings
		tests        []discovery.Test
		expectedFail bool
	}{{
		name: "Namesapce exists",
		config: model.Configuration{
			Namespace: v1alpha2.NamespaceOptions{
				Name: "default",
			},
		},
		client: &fake.FakeClient{
			GetFn: func(ctx context.Context, call int, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
				return nil
			},
		},
		clock:        clock.RealClock{},
		bindings:     apis.NewBindings(),
		tests:        []discovery.Test{},
		expectedFail: false,
	}, {
		name: "Namesapce doesn't exists",
		config: model.Configuration{
			Namespace: v1alpha2.NamespaceOptions{
				Name: "chain-saw",
			},
		},
		client: &fake.FakeClient{
			GetFn: func(ctx context.Context, call int, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
				return errors.NewNotFound(v1alpha1.Resource("Namespace"), "chain-saw")
			},
			CreateFn: func(ctx context.Context, call int, obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
				return nil
			},
			DeleteFn: func(ctx context.Context, call int, obj ctrlclient.Object, opts ...ctrlclient.DeleteOption) error {
				return nil
			},
		},
		clock:        clock.RealClock{},
		bindings:     apis.NewBindings(),
		tests:        []discovery.Test{},
		expectedFail: false,
	}, {
		name: "Namesapce not found with error",
		config: model.Configuration{
			Namespace: v1alpha2.NamespaceOptions{
				Name: "chain-saw",
			},
		},
		client: &fake.FakeClient{
			GetFn: func(ctx context.Context, call int, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
				return errors.NewBadRequest("failed to get namespace")
			},
			CreateFn: func(ctx context.Context, call int, obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
				return nil
			},
		},
		clock:        clock.RealClock{},
		bindings:     apis.NewBindings(),
		tests:        []discovery.Test{},
		expectedFail: true,
	}, {
		name: "Namesapce doesn't exists and can't be created",
		config: model.Configuration{
			Namespace: v1alpha2.NamespaceOptions{
				Name: "chain-saw",
			},
		},
		client: &fake.FakeClient{
			GetFn: func(ctx context.Context, call int, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
				return errors.NewNotFound(v1alpha1.Resource("Namespace"), "chain-saw")
			},
			CreateFn: func(ctx context.Context, call int, obj ctrlclient.Object, opts ...ctrlclient.CreateOption) error {
				return errors.NewBadRequest("failed to create namespace")
			},
		},
		clock:        clock.RealClock{},
		bindings:     apis.NewBindings(),
		tests:        []discovery.Test{},
		expectedFail: true,
	}, {
		name: "Success",
		config: model.Configuration{
			Namespace: v1alpha2.NamespaceOptions{
				Name: "default",
			},
		},
		client: &fake.FakeClient{
			GetFn: func(ctx context.Context, call int, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
				return nil
			},
		},
		clock:    clock.RealClock{},
		bindings: apis.NewBindings(),
		tests: []discovery.Test{
			{
				Err:      nil,
				BasePath: "fakePath",
				Test:     &model.Test{},
			},
		},
		expectedFail: false,
	}, {
		name: "Fail",
		config: model.Configuration{
			Namespace: v1alpha2.NamespaceOptions{
				Name: "default",
			},
		},
		client: &fake.FakeClient{
			GetFn: func(ctx context.Context, call int, key ctrlclient.ObjectKey, obj ctrlclient.Object, opts ...ctrlclient.GetOption) error {
				return nil
			},
		},
		clock:    clock.RealClock{},
		bindings: apis.NewBindings(),
		tests: []discovery.Test{
			{
				Err:      errors.NewBadRequest("failed to get test"),
				BasePath: "fakePath",
				Test:     nil,
			},
		},
		expectedFail: true,
	}}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registry := mocks.Registry{}
			if tc.client != nil {
				registry.Client = tc.client
			}
			nt := &testing.MockT{}
			ctx := context.Background()
			tcontext := enginecontext.MakeContext(apis.NewBindings(), registry)
			runner := runner{
				clock: tc.clock,
			}
			runner.runTests(ctx, nt, tc.config.Namespace, tcontext, tc.tests...)
			if tc.expectedFail {
				assert.True(t, nt.FailedVar, "expected an error but got none")
			} else {
				assert.False(t, nt.FailedVar, "expected no error but got one")
			}
		})
	}
}
