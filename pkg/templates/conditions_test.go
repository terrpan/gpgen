package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test constants to avoid duplicate literal warnings
const (
	// Input condition strings
	testContainerEnabledInput = "container.enabled"
	testContainerEnabledTemplate = "{{ .Inputs.container.enabled }}"
	testContainerBuildAlwaysBuildTemplate = "{{ .Inputs.container.build.alwaysBuild }}"
	testContainerBuildOnPRTemplate = "{{ .Inputs.container.build.onPR }}"
	testContainerBuildOnProductionTemplate = "{{ .Inputs.container.build.onProduction }}"
	testContainerPushEnabledTemplate = "{{ .Inputs.container.push.enabled }}"
	testContainerPushAlwaysPushTemplate = "{{ .Inputs.container.push.alwaysPush }}"
	testContainerPushOnProductionTemplate = "{{ .Inputs.container.push.onProduction }}"
	testSecurityTrivyEnabledInput = "security.trivy.enabled"
	testSecurityTrivyEnabledTemplate = "{{ .Inputs.security.trivy.enabled }}"
	testSecurityTrivyEnabledWithAlwaysTemplate = "{{ .Inputs.security.trivy.enabled }} && always()"

	// GitHub event condition strings
	testEventPushCondition = "github.event_name == 'push'"
	testEventReleaseCondition = "github.event_name == 'release'"
	testEventPullRequestCondition = "github.event_name == 'pull_request'"

	// GitHub ref condition strings
	testRefTagsStartsWithCondition = "startsWith(github.ref, 'refs/tags/')"

	// Common event names for testing
	testEventPush = "push"
	testEventRelease = "release"

	// Ref patterns for testing
	testRefTagsPrefix = "refs/tags/"
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
		cb := NewConditionBuilder().WithInputCondition(testContainerEnabledInput)
		assert.Equal(t, testContainerEnabledTemplate, cb.And())
		assert.Equal(t, testContainerEnabledTemplate, cb.Or())
	})

	t.Run("multiple conditions with AND", func(t *testing.T) {
		cb := NewConditionBuilder().
			WithInputCondition(testContainerEnabledInput).
			WithEventEquals(testEventPush)
		expected := testContainerEnabledTemplate + " && " + testEventPushCondition
		assert.Equal(t, expected, cb.And())
	})

	t.Run("multiple conditions with OR", func(t *testing.T) {
		cb := NewConditionBuilder().
			WithEventEquals(testEventPush).
			WithEventEquals(testEventRelease)
		expected := "(" + testEventPushCondition + " || " + testEventReleaseCondition + ")"
		assert.Equal(t, expected, cb.Or())
	})

	t.Run("ref starts with condition", func(t *testing.T) {
		cb := NewConditionBuilder().WithRefStartsWith(testRefTagsPrefix)
		assert.Equal(t, testRefTagsStartsWithCondition, cb.And())
	})

	t.Run("always condition", func(t *testing.T) {
		cb := NewConditionBuilder().
			WithInputCondition(testSecurityTrivyEnabledInput).
			WithAlways()
		expected := testSecurityTrivyEnabledWithAlwaysTemplate
		assert.Equal(t, expected, cb.And())
	})

	t.Run("custom condition", func(t *testing.T) {
		cb := NewConditionBuilder().WithCustomCondition("custom_function()")
		assert.Equal(t, "custom_function()", cb.And())
	})

	t.Run("complex nested conditions", func(t *testing.T) {
		innerCondition := NewConditionBuilder().
			WithEventEquals(testEventPush).
			WithRefStartsWith(testRefTagsPrefix).
			And()

		outerCondition := NewConditionBuilder().
			WithInputCondition(testContainerEnabledInput).
			WithCustomCondition(innerCondition).
			And()

		expected := testContainerEnabledTemplate + " && " + testEventPushCondition + " && " + testRefTagsStartsWithCondition
		assert.Equal(t, expected, outerCondition)
	})
}

func TestContainerConditions(t *testing.T) {
	t.Run("build condition structure", func(t *testing.T) {
		condition := ContainerCond.BuildCondition()

		// Should contain all the main components
		assert.Contains(t, condition, testContainerEnabledTemplate)
		assert.Contains(t, condition, testContainerBuildAlwaysBuildTemplate)
		assert.Contains(t, condition, testContainerBuildOnPRTemplate)
		assert.Contains(t, condition, testContainerBuildOnProductionTemplate)
		assert.Contains(t, condition, testEventPullRequestCondition)
		assert.Contains(t, condition, testEventPushCondition)
		assert.Contains(t, condition, testRefTagsStartsWithCondition)
		assert.Contains(t, condition, testEventReleaseCondition)

		// Should be a well-formed condition (no dangling operators)
		assert.NotContains(t, condition, " && )")
		assert.NotContains(t, condition, "( || ")
		assert.NotContains(t, condition, " || )")
	})

	t.Run("push condition structure", func(t *testing.T) {
		condition := ContainerCond.PushCondition()

		// Should contain all the main components
		assert.Contains(t, condition, testContainerEnabledTemplate)
		assert.Contains(t, condition, testContainerPushEnabledTemplate)
		assert.Contains(t, condition, testContainerPushAlwaysPushTemplate)
		assert.Contains(t, condition, testContainerPushOnProductionTemplate)
		assert.Contains(t, condition, testEventPushCondition)
		assert.Contains(t, condition, testRefTagsStartsWithCondition)
		assert.Contains(t, condition, testEventReleaseCondition)

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
			testContainerEnabledTemplate,
			testContainerBuildAlwaysBuildTemplate,
			testContainerBuildOnPRTemplate,
			testEventPullRequestCondition,
			testContainerBuildOnProductionTemplate,
			testEventPushCondition,
			testRefTagsStartsWithCondition,
			testEventReleaseCondition,
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
			testContainerEnabledTemplate,
			testContainerPushEnabledTemplate,
			testContainerPushAlwaysPushTemplate,
			testContainerPushOnProductionTemplate,
			testEventPushCondition,
			testRefTagsStartsWithCondition,
			testEventReleaseCondition,
		}

		for _, part := range expectedParts {
			assert.Contains(t, condition, part, "Push condition should contain: %s", part)
		}
	})
}

func TestSecurityConditions(t *testing.T) {
	t.Run("trivy scan condition", func(t *testing.T) {
		condition := SecurityCond.TrivyScanCondition()
		assert.Equal(t, testSecurityTrivyEnabledTemplate, condition)
	})

	t.Run("trivy upload condition", func(t *testing.T) {
		condition := SecurityCond.TrivyUploadCondition()
		expected := testSecurityTrivyEnabledWithAlwaysTemplate
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
			WithEventEquals(testEventPush).
			WithRefStartsWith(testRefTagsPrefix).
			WithAlways().
			WithCustomCondition("custom()")

		assert.NotNil(t, cb)
		assert.IsType(t, &ConditionBuilder{}, cb)
	})

	t.Run("reusable builder", func(t *testing.T) {
		// Test that builder can be reused for different output formats
		cb := NewConditionBuilder().
			WithEventEquals(testEventPush).
			WithEventEquals(testEventRelease)

		andResult := cb.And()
		orResult := cb.Or()

		assert.Equal(t, testEventPushCondition+" && "+testEventReleaseCondition, andResult)
		assert.Equal(t, "("+testEventPushCondition+" || "+testEventReleaseCondition+")", orResult)
	})
}

func TestRealWorldConditionExamples(t *testing.T) {
	t.Run("production deployment condition", func(t *testing.T) {
		// Simulate a real-world condition for production deployments
		condition := NewConditionBuilder().
			WithInputCondition("deploy.enabled").
			WithCustomCondition(
				NewConditionBuilder().
					WithEventEquals(testEventPush).
					WithRefStartsWith(testRefTagsPrefix).
					And(),
			).
			WithEventEquals(testEventRelease).
			Or()

		result := NewConditionBuilder().
			WithCustomCondition(condition).
			And()

		assert.Contains(t, result, "{{ .Inputs.deploy.enabled }}")
		assert.Contains(t, result, testEventPushCondition)
		assert.Contains(t, result, testRefTagsStartsWithCondition)
		assert.Contains(t, result, testEventReleaseCondition)
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
