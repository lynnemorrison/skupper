package kube

import (
	"testing"
	"time"

	"github.com/skupperproject/skupper/internal/cmd/skupper/common"
	"github.com/skupperproject/skupper/internal/cmd/skupper/common/utils"

	fakeclient "github.com/skupperproject/skupper/internal/kube/client/fake"
	"github.com/skupperproject/skupper/pkg/apis/skupper/v1alpha1"
	"gotest.tools/assert"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCmdConnectorCreate_ValidateInput(t *testing.T) {
	type test struct {
		name           string
		args           []string
		flags          common.CommandConnectorCreateFlags
		k8sObjects     []runtime.Object
		skupperObjects []runtime.Object
		expectedErrors []string
	}

	testTable := []test{
		{
			name: "connector is not created because there is already the same connector in the namespace",
			args: []string{"my-connector", "8080"},
			flags: common.CommandConnectorCreateFlags{
				Selector: "backend",
				Timeout:  1 * time.Minute,
			},
			skupperObjects: []runtime.Object{
				&v1alpha1.Connector{
					ObjectMeta: v1.ObjectMeta{
						Name:      "my-connector",
						Namespace: "test",
					},
					Spec: v1alpha1.ConnectorSpec{
						Port:     8080,
						Type:     "tcp",
						Host:     "test",
						Selector: "app=mySelector",
					},
					Status: v1alpha1.ConnectorStatus{
						Status: v1alpha1.Status{
							Conditions: []v1.Condition{
								{
									Type:   "Configured",
									Status: "True",
								},
							},
						},
					},
				},
			},
			expectedErrors: []string{"there is already a connector my-connector created for namespace test"},
		},
		{
			name: "connector name and port are not specified",
			args: []string{},
			flags: common.CommandConnectorCreateFlags{
				Selector: "backend",
				Timeout:  1 * time.Minute,
			},
			expectedErrors: []string{"connector name and port must be configured"},
		},
		{
			name: "connector name empty",
			args: []string{"", "8090"},
			flags: common.CommandConnectorCreateFlags{
				Selector: "backend",
				Timeout:  1 * time.Minute,
			},
			expectedErrors: []string{"connector name must not be empty"},
		},
		{
			name: "connector port empty",
			args: []string{"my-name-port-empty", ""},
			flags: common.CommandConnectorCreateFlags{
				Selector: "backend",
				Timeout:  1 * time.Minute,
			},
			expectedErrors: []string{"connector port must not be empty"},
		},
		{
			name: "connector port not positive",
			args: []string{"my-port-positive", "-45"},
			flags: common.CommandConnectorCreateFlags{
				Selector: "backend",
				Timeout:  1 * time.Minute,
			},
			expectedErrors: []string{"connector port is not valid: value is not positive"},
		},
		{
			name: "connector name and port are not specified",
			args: []string{},
			flags: common.CommandConnectorCreateFlags{
				Selector: "backend",
				Timeout:  1 * time.Minute,
			},
			expectedErrors: []string{"connector name and port must be configured"},
		},
		{
			name: "connector port is not specified",
			args: []string{"my-name"},
			flags: common.CommandConnectorCreateFlags{
				Selector: "backend",
				Timeout:  1 * time.Minute,
			},
			expectedErrors: []string{"connector name and port must be configured"},
		},
		{
			name: "more than two arguments are specified",
			args: []string{"my", "connector", "8080"},
			flags: common.CommandConnectorCreateFlags{
				Selector: "backend",
				Timeout:  1 * time.Minute,
			},
			expectedErrors: []string{"only two arguments are allowed for this command"},
		},
		{
			name: "connector name is not valid.",
			args: []string{"my new connector", "8080"},
			flags: common.CommandConnectorCreateFlags{
				Selector: "backend",
				Timeout:  1 * time.Minute,
			},
			expectedErrors: []string{"connector name is not valid: value does not match this regular expression: ^[a-z0-9]([-a-z0-9]*[a-z0-9])*(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])*)*$"},
		},
		{
			name: "port is not valid.",
			args: []string{"my-connector-port", "abcd"},
			flags: common.CommandConnectorCreateFlags{
				Selector: "backend",
				Timeout:  1 * time.Minute,
			},
			expectedErrors: []string{"connector port is not valid: strconv.Atoi: parsing \"abcd\": invalid syntax"},
		},
		{
			name: "connector type is not valid",
			args: []string{"my-connector-type", "8080"},
			flags: common.CommandConnectorCreateFlags{
				ConnectorType: "not-valid",
				Timeout:       1 * time.Minute,
				Selector:      "backend",
			},
			expectedErrors: []string{
				"connector type is not valid: value not-valid not allowed. It should be one of this options: [tcp]"},
		},
		{
			name: "routing key is not valid",
			args: []string{"my-connector-rk", "8080"},
			flags: common.CommandConnectorCreateFlags{
				RoutingKey: "not-valid$",
				Timeout:    1 * time.Minute,
				Selector:   "backend",
			},
			expectedErrors: []string{
				"routing key is not valid: value does not match this regular expression: ^[a-z0-9]([-a-z0-9]*[a-z0-9])*(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])*)*$"},
		},
		{
			name: "tls-secret does not exist",
			args: []string{"my-connector-tls", "8080"},
			flags: common.CommandConnectorCreateFlags{
				TlsSecret: "not-valid",
				Timeout:   1 * time.Minute,
				Selector:  "backend",
			},
			expectedErrors: []string{"tls-secret is not valid: does not exist"},
		},
		{
			name: "workload is not valid",
			args: []string{"bad-workload", "1234"},
			flags: common.CommandConnectorCreateFlags{
				Workload: "@345",
				Timeout:  1 * time.Minute,
			},
			expectedErrors: []string{
				"workload is not valid: value does not match this regular expression: ^[A-Za-z0-9=:./-]+$"},
		},
		{
			name: "selector is not valid",
			args: []string{"bad-selector", "1234"},
			flags: common.CommandConnectorCreateFlags{
				Selector: "@#$%",
				Timeout:  20 * time.Second,
			},
			expectedErrors: []string{
				"selector is not valid: value does not match this regular expression: ^[A-Za-z0-9=:./-]+$"},
		},
		{
			name: "timeout is not valid",
			args: []string{"bad-timeout", "8080"},
			flags: common.CommandConnectorCreateFlags{
				Workload: "workload",
				Timeout:  0 * time.Second,
			},
			expectedErrors: []string{"timeout is not valid"},
		},
		{
			name: "output is not valid",
			args: []string{"bad-output", "1234"},
			flags: common.CommandConnectorCreateFlags{
				Workload: "workload",
				Timeout:  1 * time.Second,
				Output:   "not-supported",
			},
			expectedErrors: []string{
				"output type is not valid: value not-supported not allowed. It should be one of this options: [json yaml]"},
		},
		{
			name: "selector/host",
			args: []string{"selector", "1234"},
			flags: common.CommandConnectorCreateFlags{
				Timeout:  1 * time.Second,
				Output:   "json",
				Selector: "app=test",
				Host:     "test",
			},
			expectedErrors: []string{
				"If host is configured, cannot configure workload or selector",
				"If selector is configured, cannot configure workload or host"},
		},
		{
			name: "workload/host",
			args: []string{"workload", "1234"},
			flags: common.CommandConnectorCreateFlags{
				Timeout:  1 * time.Second,
				Output:   "json",
				Workload: "deployment/test",
				Host:     "test",
			},
			expectedErrors: []string{
				"If host is configured, cannot configure workload or selector",
				"If workload is configured, cannot configure selector or host"},
		},
		{
			name: "flags all valid",
			args: []string{"my-connector-flags", "8080"},
			flags: common.CommandConnectorCreateFlags{
				RoutingKey:      "routingkeyname",
				TlsSecret:       "secretname",
				ConnectorType:   "tcp",
				IncludeNotReady: true,
				Timeout:         30 * time.Second,
				Output:          "json",
			},
			k8sObjects: []runtime.Object{
				&v12.Secret{
					ObjectMeta: v1.ObjectMeta{
						Name:      "secretname",
						Namespace: "test",
					},
				},
			},
			expectedErrors: []string{},
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {

			command, err := newCmdConnectorCreateWithMocks("test", test.k8sObjects, test.skupperObjects, "")
			assert.Assert(t, err)

			command.Flags = &test.flags

			actualErrors := command.ValidateInput(test.args)

			actualErrorsMessages := utils.ErrorsToMessages(actualErrors)

			assert.DeepEqual(t, actualErrorsMessages, test.expectedErrors)

		})
	}
}

func TestCmdConnectorCreate_InputToOptions(t *testing.T) {

	type test struct {
		name                  string
		flags                 common.CommandConnectorCreateFlags
		Connectorname         string
		expectedTlsSecret     string
		expectedHost          string
		expectedSelector      string
		expectedRoutingKey    string
		expectedConnectorType string
		expectedOutput        string
		expectedTimeout       time.Duration
	}

	testTable := []test{
		{
			name:                  "test1",
			flags:                 common.CommandConnectorCreateFlags{"backend", "", "app=backend", "secret", "tcp", true, "", 2 * time.Second, "json"},
			expectedTlsSecret:     "secret",
			expectedHost:          "",
			expectedRoutingKey:    "backend",
			expectedTimeout:       2 * time.Second,
			expectedConnectorType: "tcp",
			expectedOutput:        "json",
			expectedSelector:      "app=backend",
		},
		{
			name:                  "test2",
			flags:                 common.CommandConnectorCreateFlags{"backend", "backend", "", "secret", "tcp", true, "", 2 * time.Second, "json"},
			expectedTlsSecret:     "secret",
			expectedHost:          "backend",
			expectedRoutingKey:    "backend",
			expectedTimeout:       2 * time.Second,
			expectedConnectorType: "tcp",
			expectedOutput:        "json",
			expectedSelector:      "",
		},
		{
			name:                  "test3",
			flags:                 common.CommandConnectorCreateFlags{"", "", "", "secret", "tcp", false, "", 3 * time.Second, "yaml"},
			expectedTlsSecret:     "secret",
			expectedHost:          "",
			expectedRoutingKey:    "test3",
			expectedTimeout:       3 * time.Second,
			expectedConnectorType: "tcp",
			expectedOutput:        "yaml",
			expectedSelector:      "app=test3",
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {

			cmd, err := newCmdConnectorCreateWithMocks("test", nil, nil, "")
			assert.Assert(t, err)

			cmd.Flags = &test.flags
			cmd.name = test.name

			cmd.InputToOptions()

			assert.Check(t, cmd.routingKey == test.expectedRoutingKey)
			assert.Check(t, cmd.output == test.expectedOutput)
			assert.Check(t, cmd.tlsSecret == test.expectedTlsSecret)
			assert.Check(t, cmd.host == test.expectedHost)
			assert.Check(t, cmd.timeout == test.expectedTimeout)
			assert.Check(t, cmd.selector == test.expectedSelector)
			assert.Check(t, cmd.connectorType == test.expectedConnectorType)
		})
	}
}

func TestCmdConnectorCreate_Run(t *testing.T) {
	type test struct {
		name                string
		connectorName       string
		connectorPort       int
		flags               common.CommandConnectorCreateFlags
		k8sObjects          []runtime.Object
		skupperObjects      []runtime.Object
		skupperErrorMessage string
		errorMessage        string
	}

	testTable := []test{
		{
			name:          "runs ok",
			connectorName: "my-connector-ok",
			connectorPort: 8080,
			flags: common.CommandConnectorCreateFlags{
				ConnectorType:   "tcp",
				RoutingKey:      "keyname",
				TlsSecret:       "secretname",
				IncludeNotReady: true,
				Selector:        "app=backend",
				Timeout:         1 * time.Second,
			},
		},
		{
			name:          "run output json",
			connectorName: "my-connector-json",
			connectorPort: 8080,
			flags: common.CommandConnectorCreateFlags{
				ConnectorType:   "tcp",
				Host:            "hostname",
				RoutingKey:      "keyname",
				TlsSecret:       "secretname",
				IncludeNotReady: true,
				Timeout:         1 * time.Second,
				Output:          "json",
			},
		},
	}

	for _, test := range testTable {
		cmd, err := newCmdConnectorCreateWithMocks("test", test.k8sObjects, test.skupperObjects, test.skupperErrorMessage)
		assert.Assert(t, err)

		t.Run(test.name, func(t *testing.T) {

			cmd.Flags = &common.CommandConnectorCreateFlags{}
			cmd.name = test.connectorName
			cmd.port = test.connectorPort
			cmd.output = test.flags.Output
			cmd.namespace = "test"

			err := cmd.Run()
			if err != nil {
				assert.Check(t, test.errorMessage == err.Error())
			} else {
				assert.Check(t, err == nil)
			}
		})
	}
}

func TestCmdConnectorCreate_WaitUntil(t *testing.T) {
	type test struct {
		name                string
		output              string
		k8sObjects          []runtime.Object
		skupperObjects      []runtime.Object
		skupperErrorMessage string
		expectError         bool
	}

	testTable := []test{
		{
			name: "connector is not ready",
			skupperObjects: []runtime.Object{
				&v1alpha1.Connector{
					ObjectMeta: v1.ObjectMeta{
						Name:      "my-connector",
						Namespace: "test",
					},
					Status: v1alpha1.ConnectorStatus{},
				},
			},
			expectError: true,
		},
		{
			name:        "connector is not returned",
			expectError: true,
		},
		{
			name: "connector is ready",
			skupperObjects: []runtime.Object{
				&v1alpha1.Connector{
					ObjectMeta: v1.ObjectMeta{
						Name:      "my-connector",
						Namespace: "test",
					},
					Status: v1alpha1.ConnectorStatus{
						Status: v1alpha1.Status{
							Conditions: []v1.Condition{
								{
									Type:   "Configured",
									Status: "True",
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name:   "connector is ready yaml output",
			output: "yaml",
			skupperObjects: []runtime.Object{
				&v1alpha1.Connector{
					ObjectMeta: v1.ObjectMeta{
						Name:      "my-connector",
						Namespace: "test",
					},
					Status: v1alpha1.ConnectorStatus{
						Status: v1alpha1.Status{
							Conditions: []v1.Condition{
								{
									Type:   "Configured",
									Status: "True",
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
	}

	for _, test := range testTable {
		cmd, err := newCmdConnectorCreateWithMocks("test", test.k8sObjects, test.skupperObjects, test.skupperErrorMessage)
		assert.Assert(t, err)

		cmd.name = "my-connector"
		cmd.output = test.output
		cmd.timeout = 1 * time.Second
		cmd.namespace = "test"

		t.Run(test.name, func(t *testing.T) {

			err := cmd.WaitUntil()
			if test.expectError {
				assert.Check(t, err != nil)
			} else {
				assert.Assert(t, err)
			}
		})
	}
}

// --- helper methods

func newCmdConnectorCreateWithMocks(namespace string, k8sObjects []runtime.Object, skupperObjects []runtime.Object, fakeSkupperError string) (*CmdConnectorCreate, error) {

	// We make sure the interval is appropriate
	utils.SetRetryProfile(utils.TestRetryProfile)
	client, err := fakeclient.NewFakeClient(namespace, k8sObjects, skupperObjects, fakeSkupperError)
	if err != nil {
		return nil, err
	}
	cmdConnectorCreate := &CmdConnectorCreate{
		client:     client.GetSkupperClient().SkupperV1alpha1(),
		KubeClient: client.GetKubeClient(),
		namespace:  namespace,
	}
	return cmdConnectorCreate, nil
}
