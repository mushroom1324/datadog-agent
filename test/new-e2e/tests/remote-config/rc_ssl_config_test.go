// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package remoteconfig

import (
	_ "embed"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DataDog/test-infra-definitions/components/datadog/agentparams"

	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
	awshost "github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners/aws/host"
)

type sslConfigSuite struct {
	e2e.BaseSuite[environments.Host]
}

//go:embed fixtures/ssl_mismatch.yaml
var sslMismatchConfig string

func TestSslConfigSuite(t *testing.T) {
	t.Parallel()
	e2e.Run(t, &sslConfigSuite{},
		e2e.WithProvisioner(
			awshost.ProvisionerNoFakeIntake(
				awshost.WithAgentOptions(
					agentparams.WithAgentConfig(sslMismatchConfig),
				),
			),
		),
	)
}

// TestRemoteConfigSSLConfigMismatch tests the startup condition where the agent's SSL config is disabled but RC's TLS validation is not explicitly disabled
func (s *sslConfigSuite) TestRemoteConfigSSLConfigMismatch() {
	expectedLogs := []string{
		// Ensure the agent logs a warning about the SSL config mismatch
		"remote Configuration does not allow skipping TLS validation by default",
		// Ensure the remote config service stops, and the client stops because the service is no longer responding
		"remote configuration isn't enabled, disabling client",
		// Ensure the agent logs a warning about the remote config service being unable to start
		"unable to create remote config service",
	}
	assertAgentLogsEventually(s.T(), s.Env().RemoteHost, "agent", expectedLogs, 2*time.Minute, 5*time.Second)

	// Ensure the agent remains running despite the remote config service initialization failure
	// EventuallyWithT will wait for the duration of the `tick` argument before executing the assertion function,
	// so we can use a long `waitFor` and a slightly-less-long `tick` to ensure we wait 55s before performing the assertion.
	agentStayedHealthy := s.EventuallyWithT(func(c *assert.CollectT) {
		assert.True(c, s.Env().Agent.Client.IsReady())
	}, 1*time.Minute, 55*time.Second)
	assert.Truef(s.T(), agentStayedHealthy, "Agent shut down after remote config initialization failed")
}
