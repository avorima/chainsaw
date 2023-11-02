package logging

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clock "k8s.io/utils/clock/testing"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestNewLogger(t *testing.T) {
	fakeClock := clock.NewFakePassiveClock(time.Now())
	testName := "testName"
	stepName := "stepName"
	logger, ok := NewLogger(t, fakeClock, testName, stepName).(*logger)

	assert.True(t, ok, "Type assertion for *logger failed")

	assert.Equal(t, t, logger.t)
	assert.Equal(t, fakeClock, logger.clock)
	assert.Equal(t, testName, logger.test)
	assert.Equal(t, stepName, logger.step)
	assert.Nil(t, logger.resource)
}

func TestLog(t *testing.T) {
	fakeClock := clock.NewFakePassiveClock(time.Now())
	mockT := &TestLogger{}

	fakeLogger := NewLogger(mockT, fakeClock, "testName", "stepName").(*logger)

	testCases := []struct {
		name           string
		resource       ctrlclient.Object
		operation      string
		args           []interface{}
		expectContains []string
	}{
		{
			name:      "Without Resource",
			resource:  nil,
			operation: "OPERATION",
			args:      []interface{}{"arg1", "arg2"},
			expectContains: []string{
				"testName", "stepName", "OPERATION", "arg1", "arg2",
			},
		},
		{
			name: "With Resource",
			resource: &testResource{
				name:      "testResource",
				namespace: "default",
				gvk:       schema.GroupVersionKind{Group: "testGroup", Version: "v1", Kind: "testKind"},
			},
			operation: "OPERATION",
			args:      []interface{}{"arg1", "arg2"},
			expectContains: []string{
				"testName", "stepName", "OPERATION", "default/testResource", "testGroup/v1/testKind", "arg1", "arg2",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.resource != nil {
				fakeLogger = fakeLogger.WithResource(tt.resource).(*logger)
			}

			fakeLogger.Log(tt.operation, nil, tt.args...)
			for _, exp := range tt.expectContains {
				found := false
				for _, msg := range mockT.messages {
					if strings.Contains(msg, exp) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find '%s' in logs, but didn't. Logs: %v", exp, mockT.messages)
			}
			mockT.messages = []string{}
		})
	}
}

func TestWithResource(t *testing.T) {
	testCases := []struct {
		name      string
		resource  ctrlclient.Object
		expectNil bool
	}{
		{"Valid Resource", &testResource{name: "testResource"}, false},
		{"Nil Resource", nil, true},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			fakeClock := clock.NewFakePassiveClock(time.Now())
			fakeLogger := logger{
				t:     t,
				clock: fakeClock,
				test:  "testName",
				step:  "stepName",
			}

			newLogger := fakeLogger.WithResource(tt.resource).(*logger)

			if tt.expectNil {
				assert.Nil(t, newLogger.resource, "Expected resource to be nil in the logger")
			} else {
				assert.NotNil(t, newLogger.resource, "Expected resource to not be nil in the logger")
				assert.Equal(t, tt.resource, newLogger.resource, "Expected correct resource to be set in the logger")
			}

			assert.Equal(t, fakeLogger.t, newLogger.t, "Expected testing.T to remain the same")
			assert.Equal(t, fakeLogger.clock, newLogger.clock, "Expected clock to remain the same")
			assert.Equal(t, fakeLogger.test, newLogger.test, "Expected test name to remain the same")
			assert.Equal(t, fakeLogger.step, newLogger.step, "Expected step name to remain the same")
		})
	}
}