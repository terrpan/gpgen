package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyDefaults_PartialPushConfig(t *testing.T) {
	p := NewInputProcessor()
	raw := map[string]interface{}{
		"container": map[string]interface{}{
			"push": map[string]interface{}{
				"onProduction": false,
			},
		},
	}

	inputs, err := p.ProcessInputs(raw)
	require.NoError(t, err)

	assert.True(t, inputs.Container.Push.Enabled)
	assert.False(t, inputs.Container.Push.OnProduction)
}

func TestApplyDefaults_PartialBuildConfig(t *testing.T) {
	p := NewInputProcessor()
	raw := map[string]interface{}{
		"container": map[string]interface{}{
			"build": map[string]interface{}{
				"onPR": false,
			},
		},
	}

	inputs, err := p.ProcessInputs(raw)
	require.NoError(t, err)

	def := DefaultContainerConfig()

	assert.False(t, inputs.Container.Build.OnPR)
	assert.Equal(t, def.Build.OnProduction, inputs.Container.Build.OnProduction)
	assert.Equal(t, def.Build.AlwaysBuild, inputs.Container.Build.AlwaysBuild)
	assert.Equal(t, def.Build.AlwaysPush, inputs.Container.Build.AlwaysPush)
}

func TestApplyDefaults_DisablePush(t *testing.T) {
	p := NewInputProcessor()
	raw := map[string]interface{}{
		"container": map[string]interface{}{
			"push": map[string]interface{}{
				"enabled": false,
			},
		},
	}

	inputs, err := p.ProcessInputs(raw)
	require.NoError(t, err)

	def := DefaultContainerConfig()

	assert.False(t, inputs.Container.Push.Enabled)
	assert.Equal(t, def.Push.OnProduction, inputs.Container.Push.OnProduction)
}
