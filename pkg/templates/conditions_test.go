package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHubActionVersions(t *testing.T) {
	t.Run("checkout version", func(t *testing.T) {
		assert.Equal(t, "actions/checkout@v4", GitHubActionVersions.Checkout)
	})

	t.Run("setup node version", func(t *testing.T) {
		assert.Equal(t, "actions/setup-node@v4", GitHubActionVersions.SetupNode)
	})

	t.Run("setup go version", func(t *testing.T) {
		assert.Equal(t, "actions/setup-go@v4", GitHubActionVersions.SetupGo)
	})

	t.Run("setup python version", func(t *testing.T) {
		assert.Equal(t, "actions/setup-python@v4", GitHubActionVersions.SetupPython)
	})

	t.Run("docker actions versions", func(t *testing.T) {
		assert.Equal(t, "docker/setup-buildx-action@v3", GitHubActionVersions.DockerSetupBuildx)
		assert.Equal(t, "docker/login-action@v3", GitHubActionVersions.DockerLogin)
		assert.Equal(t, "docker/build-push-action@v5", GitHubActionVersions.DockerBuildPush)
	})

	t.Run("security actions versions", func(t *testing.T) {
		assert.Equal(t, "github/codeql-action/upload-sarif@v3", GitHubActionVersions.CodeQLUploadSARIF)
		assert.Equal(t, "aquasecurity/trivy-action@master", GitHubActionVersions.TrivyAction)
	})
}

func TestGitHubPlaceholders(t *testing.T) {
	t.Run("actor placeholder", func(t *testing.T) {
		assert.Equal(t, "GITHUB_ACTOR_PLACEHOLDER", GitHubPlaceholders.ActorPlaceholder)
	})

	t.Run("token placeholder", func(t *testing.T) {
		assert.Equal(t, "GITHUB_TOKEN_PLACEHOLDER", GitHubPlaceholders.TokenPlaceholder)
	})
}

func TestConditionBuilder(t *testing.T) {
	t.Run("empty builder", func(t *testing.T) {
		cb := NewConditionBuilder()
		assert.Equal(t, "", cb.And())
		assert.Equal(t, "", cb.Or())
	})

	t.Run("single condition", func(t *testing.T) {
		cb := NewConditionBuilder().WithInputCondition("container.enabled")
		assert.Equal(t, "{{ .Inputs.container.enabled }}", cb.And())
		assert.Equal(t, "{{ .Inputs.container.enabled }}", cb.Or())
	})

	t.Run("multiple conditions with AND", func(t *testing.T) {
		cb := NewConditionBuilder().
			WithInputCondition("container.enabled").
			WithEventEquals("push")
		expected := "{{ .Inputs.container.enabled }} && github.event_name == 'push'"
		assert.Equal(t, expected, cb.And())
	})

	t.Run("multiple conditions with OR", func(t *testing.T) {
		cb := NewConditionBuilder().
			WithEventEquals("push").
			WithEventEquals("release")
		expected := "(github.event_name == 'push' || github.event_name == 'release')"
		assert.Equal(t, expected, cb.Or())
	})

	t.Run("ref starts with condition", func(t *testing.T) {
		cb := NewConditionBuilder().WithRefStartsWith("refs/tags/")
		assert.Equal(t, "startsWith(github.ref, 'refs/tags/')", cb.And())
	})

	t.Run("always condition", func(t *testing.T) {
		cb := NewConditionBuilder().
			WithInputCondition("security.trivy.enabled").
			WithAlways()
		expected := "{{ .Inputs.security.trivy.enabled }} && always()"
		assert.Equal(t, expected, cb.And())
	})

	t.Run("custom condition", func(t *testing.T) {
		cb := NewConditionBuilder().WithCustomCondition("custom_function()")
		assert.Equal(t, "custom_function()", cb.And())
	})

	t.Run("complex nested conditions", func(t *testing.T) {
		innerCondition := NewConditionBuilder().
			WithEventEquals("push").
			WithRefStartsWith("refs/tags/").
			And()

		outerCondition := NewConditionBuilder().
			WithInputCondition("container.enabled").
			WithCustomCondition(innerCondition).
			And()

		expected := "{{ .Inputs.container.enabled }} && github.event_name == 'push' && startsWith(github.ref, 'refs/tags/')"
		assert.Equal(t, expected, outerCondition)
	})
}

func TestContainerConditions(t *testing.T) {
	t.Run("build condition structure", func(t *testing.T) {
		condition := ContainerCond.BuildCondition()

		// Should contain all the main components
		assert.Contains(t, condition, "{{ .Inputs.container.enabled }}")
		assert.Contains(t, condition, "{{ .Inputs.container.build.alwaysBuild }}")
		assert.Contains(t, condition, "{{ .Inputs.container.build.onPR }}")
		assert.Contains(t, condition, "{{ .Inputs.container.build.onProduction }}")
		assert.Contains(t, condition, "github.event_name == 'pull_request'")
		assert.Contains(t, condition, "github.event_name == 'push'")
		assert.Contains(t, condition, "startsWith(github.ref, 'refs/tags/')")
		assert.Contains(t, condition, "github.event_name == 'release'")

		// Should be a well-formed condition (no dangling operators)
		assert.NotContains(t, condition, " && )")
		assert.NotContains(t, condition, "( || ")
		assert.NotContains(t, condition, " || )")
	})

	t.Run("push condition structure", func(t *testing.T) {
		condition := ContainerCond.PushCondition()

		// Should contain all the main components
		assert.Contains(t, condition, "{{ .Inputs.container.enabled }}")
		assert.Contains(t, condition, "{{ .Inputs.container.push.enabled }}")
		assert.Contains(t, condition, "{{ .Inputs.container.push.alwaysPush }}")
		assert.Contains(t, condition, "{{ .Inputs.container.push.onProduction }}")
		assert.Contains(t, condition, "github.event_name == 'push'")
		assert.Contains(t, condition, "startsWith(github.ref, 'refs/tags/')")
		assert.Contains(t, condition, "github.event_name == 'release'")

		// Should be a well-formed condition
		assert.NotContains(t, condition, " && )")
		assert.NotContains(t, condition, "( || ")
		assert.NotContains(t, condition, " || )")
	})

	t.Run("build condition matches expected pattern", func(t *testing.T) {
		condition := ContainerCond.BuildCondition()

		// The condition should follow the pattern:
		// container.enabled && (alwaysBuild || (onPR && pull_request) || (onProduction && (push+tags || release)))
		expectedParts := []string{
			"{{ .Inputs.container.enabled }}",
			"{{ .Inputs.container.build.alwaysBuild }}",
			"{{ .Inputs.container.build.onPR }}",
			"github.event_name == 'pull_request'",
			"{{ .Inputs.container.build.onProduction }}",
			"github.event_name == 'push'",
			"startsWith(github.ref, 'refs/tags/')",
			"github.event_name == 'release'",
		}

		for _, part := range expectedParts {
			assert.Contains(t, condition, part, "Build condition should contain: %s", part)
		}
	})

	t.Run("push condition matches expected pattern", func(t *testing.T) {
		condition := ContainerCond.PushCondition()

		// The condition should follow the pattern:
		// container.enabled && push.enabled && (alwaysPush || (onProduction && (push+tags || release)))
		expectedParts := []string{
			"{{ .Inputs.container.enabled }}",
			"{{ .Inputs.container.push.enabled }}",
			"{{ .Inputs.container.push.alwaysPush }}",
			"{{ .Inputs.container.push.onProduction }}",
			"github.event_name == 'push'",
			"startsWith(github.ref, 'refs/tags/')",
			"github.event_name == 'release'",
		}

		for _, part := range expectedParts {
			assert.Contains(t, condition, part, "Push condition should contain: %s", part)
		}
	})
}

func TestSecurityConditions(t *testing.T) {
	t.Run("trivy scan condition", func(t *testing.T) {
		condition := SecurityCond.TrivyScanCondition()
		assert.Equal(t, "{{ .Inputs.security.trivy.enabled }}", condition)
	})

	t.Run("trivy upload condition", func(t *testing.T) {
		condition := SecurityCond.TrivyUploadCondition()
		expected := "{{ .Inputs.security.trivy.enabled }} && always()"
		assert.Equal(t, expected, condition)
	})
}

func TestEventConstants(t *testing.T) {
	t.Run("event names", func(t *testing.T) {
		assert.Equal(t, "pull_request", EventPullRequest)
		assert.Equal(t, "push", EventPush)
		assert.Equal(t, "release", EventRelease)
	})

	t.Run("ref patterns", func(t *testing.T) {
		assert.Equal(t, "refs/tags/", RefTagsPrefix)
		assert.Equal(t, "refs/heads/main", RefMainBranch)
	})

	t.Run("github context variables", func(t *testing.T) {
		assert.Equal(t, "github.event_name", GitHubEventName)
		assert.Equal(t, "github.ref", GitHubRef)
	})
}

func TestConditionBuilderChaining(t *testing.T) {
	t.Run("method chaining", func(t *testing.T) {
		// Test that all methods return *ConditionBuilder for chaining
		cb := NewConditionBuilder().
			WithInputCondition("test.input").
			WithEventEquals("push").
			WithRefStartsWith("refs/tags/").
			WithAlways().
			WithCustomCondition("custom()")

		assert.NotNil(t, cb)
		assert.IsType(t, &ConditionBuilder{}, cb)
	})

	t.Run("reusable builder", func(t *testing.T) {
		// Test that builder can be reused for different output formats
		cb := NewConditionBuilder().
			WithEventEquals("push").
			WithEventEquals("release")

		andResult := cb.And()
		orResult := cb.Or()

		assert.Equal(t, "github.event_name == 'push' && github.event_name == 'release'", andResult)
		assert.Equal(t, "(github.event_name == 'push' || github.event_name == 'release')", orResult)
	})
}

func TestRealWorldConditionExamples(t *testing.T) {
	t.Run("production deployment condition", func(t *testing.T) {
		// Simulate a real-world condition for production deployments
		condition := NewConditionBuilder().
			WithInputCondition("deploy.enabled").
			WithCustomCondition(
				NewConditionBuilder().
					WithEventEquals("push").
					WithRefStartsWith("refs/tags/").
					And(),
			).
			WithEventEquals("release").
			Or()

		result := NewConditionBuilder().
			WithCustomCondition(condition).
			And()

		assert.Contains(t, result, "{{ .Inputs.deploy.enabled }}")
		assert.Contains(t, result, "github.event_name == 'push'")
		assert.Contains(t, result, "startsWith(github.ref, 'refs/tags/')")
		assert.Contains(t, result, "github.event_name == 'release'")
	})

	t.Run("security scan with environment condition", func(t *testing.T) {
		// Simulate a condition that only runs security scans in certain environments
		condition := NewConditionBuilder().
			WithInputCondition("security.enabled").
			WithCustomCondition(
				NewConditionBuilder().
					WithInputCondition("environment.prod").
					WithInputCondition("environment.staging").
					Or(),
			).
			And()

		assert.Contains(t, condition, "{{ .Inputs.security.enabled }}")
		assert.Contains(t, condition, "{{ .Inputs.environment.prod }}")
		assert.Contains(t, condition, "{{ .Inputs.environment.staging }}")
	})
}
