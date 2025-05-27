package templates

import (
	"fmt"
	"strings"
)

// GitHubEventConditions contains type-safe constants for GitHub event conditions
type GitHubEventConditions struct{}

// GitHub event names
const (
	EventPullRequest = "pull_request"
	EventPush        = "push"
	EventRelease     = "release"
)

// GitHub ref patterns
const (
	RefTagsPrefix = "refs/tags/"
	RefMainBranch = "refs/heads/main"
)

// GitHub context variables
const (
	GitHubEventName = "github.event_name"
	GitHubRef       = "github.ref"
)

// GitHubActionVersions contains centralized action version constants
var GitHubActionVersions = struct {
	Checkout         string
	SetupNode        string
	SetupGo          string
	SetupPython      string
	DockerSetupBuildx string
	DockerLogin      string
	DockerBuildPush  string
	CodeQLUploadSARIF string
	TrivyAction      string
}{
	Checkout:         "actions/checkout@v4",
	SetupNode:        "actions/setup-node@v4",
	SetupGo:          "actions/setup-go@v4",
	SetupPython:      "actions/setup-python@v4",
	DockerSetupBuildx: "docker/setup-buildx-action@v3",
	DockerLogin:      "docker/login-action@v3",
	DockerBuildPush:  "docker/build-push-action@v5",
	CodeQLUploadSARIF: "github/codeql-action/upload-sarif@v3",
	TrivyAction:      "aquasecurity/trivy-action@master",
}

// GitHubPlaceholders contains centralized placeholder constants
var GitHubPlaceholders = struct {
	ActorPlaceholder string
	TokenPlaceholder string
}{
	ActorPlaceholder: "GITHUB_ACTOR_PLACEHOLDER",
	TokenPlaceholder: "GITHUB_TOKEN_PLACEHOLDER",
}

// ConditionBuilder helps construct complex GitHub Actions conditional expressions
type ConditionBuilder struct {
	parts []string
}

// NewConditionBuilder creates a new condition builder
func NewConditionBuilder() *ConditionBuilder {
	return &ConditionBuilder{
		parts: make([]string, 0),
	}
}

// WithInputCondition adds an input-based condition
func (cb *ConditionBuilder) WithInputCondition(inputPath string) *ConditionBuilder {
	cb.parts = append(cb.parts, fmt.Sprintf("{{ .Inputs.%s }}", inputPath))
	return cb
}

// WithEventEquals adds an event name equality condition
func (cb *ConditionBuilder) WithEventEquals(eventName string) *ConditionBuilder {
	cb.parts = append(cb.parts, fmt.Sprintf("%s == '%s'", GitHubEventName, eventName))
	return cb
}

// WithRefStartsWith adds a ref prefix condition
func (cb *ConditionBuilder) WithRefStartsWith(prefix string) *ConditionBuilder {
	cb.parts = append(cb.parts, fmt.Sprintf("startsWith(%s, '%s')", GitHubRef, prefix))
	return cb
}

// WithAlways adds the always() function
func (cb *ConditionBuilder) WithAlways() *ConditionBuilder {
	cb.parts = append(cb.parts, "always()")
	return cb
}

// WithCustomCondition adds a custom condition string
func (cb *ConditionBuilder) WithCustomCondition(condition string) *ConditionBuilder {
	cb.parts = append(cb.parts, condition)
	return cb
}

// And combines all parts with AND operator
func (cb *ConditionBuilder) And() string {
	if len(cb.parts) == 0 {
		return ""
	}
	if len(cb.parts) == 1 {
		return cb.parts[0]
	}
	return strings.Join(cb.parts, " && ")
}

// Or combines all parts with OR operator
func (cb *ConditionBuilder) Or() string {
	if len(cb.parts) == 0 {
		return ""
	}
	if len(cb.parts) == 1 {
		return cb.parts[0]
	}
	return "(" + strings.Join(cb.parts, " || ") + ")"
}

// ContainerConditions provides pre-built condition builders for common container scenarios
type ContainerConditions struct{}

// BuildCondition creates the standard container build condition
// Covers: alwaysBuild || (onPR && pull_request) || (onProduction && (push+tags || release))
func (cc *ContainerConditions) BuildCondition() string {
	// Always build condition
	alwaysBuild := NewConditionBuilder().
		WithInputCondition("container.build.alwaysBuild").
		And()

	// Build on PR condition
	onPRCondition := NewConditionBuilder().
		WithInputCondition("container.build.onPR").
		WithEventEquals(EventPullRequest).
		And()

	// Build on production condition (tags or releases)
	productionEventCondition := NewConditionBuilder().
		WithEventEquals(EventPush).
		WithRefStartsWith(RefTagsPrefix).
		And()

	releaseCondition := NewConditionBuilder().
		WithEventEquals(EventRelease).
		And()

	productionEvents := NewConditionBuilder().
		WithCustomCondition(productionEventCondition).
		WithCustomCondition(releaseCondition).
		Or()

	onProductionCondition := NewConditionBuilder().
		WithInputCondition("container.build.onProduction").
		WithCustomCondition(productionEvents).
		And()

	// Combine all build conditions
	buildConditions := NewConditionBuilder().
		WithCustomCondition(alwaysBuild).
		WithCustomCondition(onPRCondition).
		WithCustomCondition(onProductionCondition).
		Or()

	// Add container enabled check
	return NewConditionBuilder().
		WithInputCondition("container.enabled").
		WithCustomCondition(buildConditions).
		And()
}

// PushCondition creates the standard container push condition
// Covers: push.enabled && (alwaysPush || (onProduction && (push+tags || release)))
func (cc *ContainerConditions) PushCondition() string {
	// Always push condition
	alwaysPush := NewConditionBuilder().
		WithInputCondition("container.push.alwaysPush").
		And()

	// Push on production condition (tags or releases)
	productionEventCondition := NewConditionBuilder().
		WithEventEquals(EventPush).
		WithRefStartsWith(RefTagsPrefix).
		And()

	releaseCondition := NewConditionBuilder().
		WithEventEquals(EventRelease).
		And()

	productionEvents := NewConditionBuilder().
		WithCustomCondition(productionEventCondition).
		WithCustomCondition(releaseCondition).
		Or()

	onProductionCondition := NewConditionBuilder().
		WithInputCondition("container.push.onProduction").
		WithCustomCondition(productionEvents).
		And()

	// Combine push conditions
	pushConditions := NewConditionBuilder().
		WithCustomCondition(alwaysPush).
		WithCustomCondition(onProductionCondition).
		Or()

	// Add container and push enabled checks
	return NewConditionBuilder().
		WithInputCondition("container.enabled").
		WithInputCondition("container.push.enabled").
		WithCustomCondition(pushConditions).
		And()
}

// SecurityConditions provides pre-built condition builders for security scenarios
type SecurityConditions struct{}

// TrivyScanCondition creates the standard Trivy scan condition
func (sc *SecurityConditions) TrivyScanCondition() string {
	return NewConditionBuilder().
		WithInputCondition("security.trivy.enabled").
		And()
}

// TrivyUploadCondition creates the Trivy SARIF upload condition (runs even on failure)
func (sc *SecurityConditions) TrivyUploadCondition() string {
	return NewConditionBuilder().
		WithInputCondition("security.trivy.enabled").
		WithAlways().
		And()
}

// Global instances for easy access
var (
	ContainerCond = &ContainerConditions{}
	SecurityCond  = &SecurityConditions{}
)
