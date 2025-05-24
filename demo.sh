#!/bin/bash

# GPGen Demo Script
# This script demonstrates the core functionality of GPGen

set -e

echo "🚀 GPGen Demo - Phase 2 Core Functionality"
echo "=========================================="
echo

# Clean up any existing files
rm -rf manifest.yaml .github/ 2>/dev/null || true

echo "📋 Step 1: Initialize a Node.js manifest"
./bin/gpgen init node-app demo-app
echo

echo "📝 Step 2: Validate the generated manifest"
./bin/gpgen validate manifest.yaml
echo

echo "🏗️  Step 3: Generate workflows (dry run)"
./bin/gpgen generate manifest.yaml --dry-run
echo

echo "✨ Step 4: Generate actual workflow files"
./bin/gpgen generate manifest.yaml
echo

echo "📄 Step 5: Show generated workflow structure"
find .github -name "*.yml" -exec echo "📁 {}" \; -exec head -10 {} \; -exec echo \;

echo "🧪 Step 6: Test with a complex manifest"
cat > complex-manifest.yaml << 'EOF'
apiVersion: gpgen.dev/v1
kind: Pipeline
metadata:
  name: complex-demo
  annotations:
    gpgen.dev/description: "Demo with custom steps"
spec:
  template: node-app
  inputs:
    nodeVersion: "20"
    packageManager: npm
    testCommand: npm run test:ci
  customSteps:



































echo "   ✅ Validation system robust"echo "   ✅ Environment overrides"echo "   ✅ Custom steps positioning"echo "   ✅ Workflow generation working"echo "   ✅ Template system functional" echo "   ✅ CLI commands working"echo "🎉 GPGen Phase 2 demonstration successful!"rm -rf manifest.yaml complex-manifest.yaml .github/echo "🧹 Cleaning up demo files..."echols -la .github/workflows/echo "✅ Demo complete! Generated files:"echo./bin/gpgen generate complex-manifest.yaml --environment productionecho "🔨 Generating complex workflow..."echo./bin/gpgen validate complex-manifest.yaml --strictecho "🔍 Validating complex manifest..."EOF        testCommand: npm run test:all      inputs:    production:  environments:      run: echo "Checking code quality..."      position: before:build    - name: Code Quality      run: echo "Running security scan..."      position: after:test    - name: Security Scan
