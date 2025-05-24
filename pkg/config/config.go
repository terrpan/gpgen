package config

// LanguageVersions defines the supported versions for different programming languages
var LanguageVersions = map[string][]string{
	"go": {
		"1.21",
		"1.22",
		"1.23",
		"1.24",
	},
	"node": {
		"16",
		"18",
		"20",
		"22",
	},
	"python": {
		"3.9",
		"3.10",
		"3.11",
		"3.12",
	},
}

// PackageManagers defines the supported package managers for different languages
var PackageManagers = map[string][]string{
	"node": {
		"npm",
		"yarn",
		"pnpm",
	},
	"python": {
		"pip",
		"poetry",
		"pipenv",
	},
}

// SecuritySeverityLevels defines the available Trivy severity levels
var SecuritySeverityLevels = []string{
	"CRITICAL",
	"HIGH",
	"MEDIUM",
	"LOW",
	"CRITICAL,HIGH",
	"CRITICAL,HIGH,MEDIUM",
}

// DefaultValues defines default values for common template inputs
var DefaultValues = map[string]interface{}{
	"goVersion":     "1.21",
	"nodeVersion":   "18",
	"pythonVersion": "3.11",

	"packageManager": map[string]string{
		"node":   "npm",
		"python": "pip",
	},

	"testCommand": map[string]string{
		"go":     "go test ./...",
		"node":   "npm test",
		"python": "pytest",
	},

	"buildCommand": map[string]string{
		"go":   "go build -o bin/service ./cmd/service",
		"node": "npm run build",
	},

	"lintCommand": map[string]string{
		"python": "flake8",
	},

	"requirements": map[string]string{
		"python": "requirements.txt",
	},
}
