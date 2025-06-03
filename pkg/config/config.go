package config

import "fmt"

// InputField represents a template input field name (type-safe alternative to string constants)
type InputField string

const (
	InputFieldGoVersion      InputField = "goVersion"
	InputFieldNodeVersion    InputField = "nodeVersion"
	InputFieldPythonVersion  InputField = "pythonVersion"
	InputFieldPackageManager InputField = "packageManager"
	InputFieldTestCommand    InputField = "testCommand"
	InputFieldBuildCommand   InputField = "buildCommand"
	InputFieldLintCommand    InputField = "lintCommand"
	InputFieldRequirements   InputField = "requirements"
)

// LanguageInputFields maps languages to their relevant input fields
var LanguageInputFields = map[Language][]InputField{
	LanguageGo:     {InputFieldGoVersion, InputFieldTestCommand, InputFieldBuildCommand},
	LanguageNode:   {InputFieldNodeVersion, InputFieldPackageManager, InputFieldTestCommand, InputFieldBuildCommand},
	LanguagePython: {InputFieldPythonVersion, InputFieldPackageManager, InputFieldTestCommand, InputFieldLintCommand, InputFieldRequirements},
}

// Language represents a supported programming language
type Language string

const (
	LanguageGo     Language = "go"
	LanguageNode   Language = "node"
	LanguagePython Language = "python"
)

// PackageManager represents a supported package manager
type PackageManager string

const (
	PackageManagerNpm    PackageManager = "npm"
	PackageManagerYarn   PackageManager = "yarn"
	PackageManagerPnpm   PackageManager = "pnpm"
	PackageManagerPip    PackageManager = "pip"
	PackageManagerPoetry PackageManager = "poetry"
	PackageManagerPipenv PackageManager = "pipenv"
)

// SecuritySeverity represents Trivy security severity levels
type SecuritySeverity string

const (
	SeverityCritical           SecuritySeverity = "CRITICAL"
	SeverityHigh               SecuritySeverity = "HIGH"
	SeverityMedium             SecuritySeverity = "MEDIUM"
	SeverityLow                SecuritySeverity = "LOW"
	SeverityCriticalHigh       SecuritySeverity = "CRITICAL,HIGH"
	SeverityCriticalHighMedium SecuritySeverity = "CRITICAL,HIGH,MEDIUM"
)

// LanguageConfig defines configuration for a specific programming language
type LanguageConfig struct {
	Versions        []string
	PackageManagers []PackageManager
	DefaultVersion  string
	DefaultManager  PackageManager
	DefaultTestCmd  string
	DefaultBuildCmd string
	DefaultLintCmd  string
	DefaultReqFile  string
}

// Configuration holds all typed configuration values
type Configuration struct {
	Languages map[Language]LanguageConfig
	Security  SecurityConfig
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	SeverityLevels []SecuritySeverity
	DefaultLevel   SecuritySeverity
}

// Config is the global configuration instance
var Config = Configuration{
	Languages: map[Language]LanguageConfig{
		LanguageGo: {
			Versions:        []string{"1.21", "1.22", "1.23", "1.24"},
			PackageManagers: []PackageManager{}, // Go uses modules, no package manager needed
			DefaultVersion:  "1.21",
			DefaultTestCmd:  "go test ./...",
			DefaultBuildCmd: "go build -o bin/service ./cmd/service",
		},
		LanguageNode: {
			Versions:        []string{"16", "18", "20", "22"},
			PackageManagers: []PackageManager{PackageManagerNpm, PackageManagerYarn, PackageManagerPnpm},
			DefaultVersion:  "18",
			DefaultManager:  PackageManagerNpm,
			DefaultTestCmd:  "npm test",
			DefaultBuildCmd: "npm run build",
		},
		LanguagePython: {
			Versions:        []string{"3.9", "3.10", "3.11", "3.12"},
			PackageManagers: []PackageManager{PackageManagerPip, PackageManagerPoetry, PackageManagerPipenv},
			DefaultVersion:  "3.11",
			DefaultManager:  PackageManagerPip,
			DefaultTestCmd:  "pytest",
			DefaultLintCmd:  "flake8",
			DefaultReqFile:  "requirements.txt",
		},
	},
	Security: SecurityConfig{
		SeverityLevels: []SecuritySeverity{
			SeverityCritical,
			SeverityHigh,
			SeverityMedium,
			SeverityLow,
			SeverityCriticalHigh,
			SeverityCriticalHighMedium,
		},
		DefaultLevel: SeverityCriticalHigh,
	},
}

// Legacy compatibility variables (deprecated - use Config methods instead)
// These are maintained for backward compatibility but will be removed in a future version

// LanguageVersions defines the supported versions for different programming languages
// Deprecated: Use Config.Languages[language].Versions instead
var LanguageVersions = map[string][]string{
	string(LanguageGo):     Config.Languages[LanguageGo].Versions,
	string(LanguageNode):   Config.Languages[LanguageNode].Versions,
	string(LanguagePython): Config.Languages[LanguagePython].Versions,
}

// PackageManagers defines the supported package managers for different languages
// Deprecated: Use Config.GetPackageManagerOptions(language) instead
var PackageManagers = map[string][]string{
	string(LanguageNode):   Config.GetPackageManagerOptions(LanguageNode),
	string(LanguagePython): Config.GetPackageManagerOptions(LanguagePython),
}

// SecuritySeverityLevels defines the available Trivy security severity levels
// Deprecated: Use Config.GetSecuritySeverityOptions() instead
var SecuritySeverityLevels = Config.GetSecuritySeverityOptions()

// Type-safe helper methods

// GetLanguageConfig returns the configuration for a specific language
func (c *Configuration) GetLanguageConfig(lang Language) (LanguageConfig, bool) {
	config, exists := c.Languages[lang]
	return config, exists
}

// IsValidVersion checks if a version is supported for the given language
func (c *Configuration) IsValidVersion(lang Language, version string) bool {
	config, exists := c.Languages[lang]
	if !exists {
		return false
	}

	for _, v := range config.Versions {
		if v == version {
			return true
		}
	}
	return false
}

// IsValidPackageManager checks if a package manager is supported for the given language
func (c *Configuration) IsValidPackageManager(lang Language, manager PackageManager) bool {
	config, exists := c.Languages[lang]
	if !exists {
		return false
	}

	for _, m := range config.PackageManagers {
		if m == manager {
			return true
		}
	}
	return false
}

// IsValidSecuritySeverity checks if a security severity level is valid
func (c *Configuration) IsValidSecuritySeverity(severity SecuritySeverity) bool {
	for _, s := range c.Security.SeverityLevels {
		if s == severity {
			return true
		}
	}
	return false
}

// GetDefaults returns default values for a specific language
// Deprecated: Use GetAllDefaults() for type-safe access or Defaults.GetXXX() methods
func (c *Configuration) GetDefaults(lang Language) map[string]interface{} {
	config, exists := c.Languages[lang]
	if !exists {
		return map[string]interface{}{}
	}

	defaults := map[string]interface{}{
		"version":                     config.DefaultVersion, // Keep legacy "version" key for backward compatibility
		string(InputFieldTestCommand): config.DefaultTestCmd,
	}

	if config.DefaultManager != "" {
		defaults[string(InputFieldPackageManager)] = string(config.DefaultManager)
	}

	if config.DefaultBuildCmd != "" {
		defaults[string(InputFieldBuildCommand)] = config.DefaultBuildCmd
	}

	if config.DefaultLintCmd != "" {
		defaults[string(InputFieldLintCommand)] = config.DefaultLintCmd
	}

	if config.DefaultReqFile != "" {
		defaults[string(InputFieldRequirements)] = config.DefaultReqFile
	}

	return defaults
}

// getVersionField returns the appropriate version field for a language
func getVersionField(lang Language) InputField {
	switch lang {
	case LanguageGo:
		return InputFieldGoVersion
	case LanguageNode:
		return InputFieldNodeVersion
	case LanguagePython:
		return InputFieldPythonVersion
	default:
		return InputFieldGoVersion // fallback
	}
}

// GetPackageManagerOptions returns package manager options as strings for a language
func (c *Configuration) GetPackageManagerOptions(lang Language) []string {
	config, exists := c.Languages[lang]
	if !exists {
		return []string{}
	}

	options := make([]string, len(config.PackageManagers))
	for i, mgr := range config.PackageManagers {
		options[i] = string(mgr)
	}
	return options
}

// GetSecuritySeverityOptions returns security severity options as strings
func (c *Configuration) GetSecuritySeverityOptions() []string {
	options := make([]string, len(c.Security.SeverityLevels))
	for i, level := range c.Security.SeverityLevels {
		options[i] = string(level)
	}
	return options
}

// GetValidInputFields returns the valid input fields for a specific language
func (c *Configuration) GetValidInputFields(lang Language) []InputField {
	if fields, exists := LanguageInputFields[lang]; exists {
		return fields
	}
	return []InputField{}
}

// GetDefaultValue returns the typed default value for a specific input field and language
func (c *Configuration) GetDefaultValue(inputField InputField, lang Language) (interface{}, error) {
	config, exists := c.Languages[lang]
	if !exists {
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}

	switch inputField {
	case InputFieldGoVersion:
		if lang == LanguageGo {
			return config.DefaultVersion, nil
		}
	case InputFieldNodeVersion:
		if lang == LanguageNode {
			return config.DefaultVersion, nil
		}
	case InputFieldPythonVersion:
		if lang == LanguagePython {
			return config.DefaultVersion, nil
		}
	case InputFieldPackageManager:
		if config.DefaultManager != "" {
			return string(config.DefaultManager), nil
		}
	case InputFieldTestCommand:
		return config.DefaultTestCmd, nil
	case InputFieldBuildCommand:
		return config.DefaultBuildCmd, nil
	case InputFieldLintCommand:
		return config.DefaultLintCmd, nil
	case InputFieldRequirements:
		return config.DefaultReqFile, nil
	}

	return nil, fmt.Errorf("invalid input field %s for language %s", inputField, lang)
}

// GetAllDefaults returns all default values for a language as a type-safe map
func (c *Configuration) GetAllDefaults(lang Language) map[InputField]interface{} {
	defaults := make(map[InputField]interface{})
	validFields := c.GetValidInputFields(lang)

	for _, field := range validFields {
		if value, err := c.GetDefaultValue(field, lang); err == nil && value != "" {
			defaults[field] = value
		}
	}

	return defaults
}

// ValidateTemplateInput validates input values against configuration constraints
func (c *Configuration) ValidateTemplateInput(inputField InputField, value interface{}, lang Language) error {
	// First check if the input field is valid for this language
	if !c.IsValidInputFieldForLanguage(inputField, lang) {
		return fmt.Errorf("input field %s is not valid for language %s", inputField, lang)
	}

	// Validate the input value based on field type
	switch inputField {
	case InputFieldNodeVersion, InputFieldGoVersion, InputFieldPythonVersion:
		if strVal, ok := value.(string); ok {
			if strVal == "" {
				return fmt.Errorf("%s version cannot be empty", lang)
			}
			if !c.IsValidVersion(lang, strVal) {
				versions, _ := c.GetVersionsForLanguage(lang)
				return fmt.Errorf("invalid %s version: %s. Supported versions: %v", lang, strVal, versions)
			}
		} else {
			return fmt.Errorf("version must be a string, got %T", value)
		}
	case InputFieldPackageManager:
		if strVal, ok := value.(string); ok {
			if strVal == "" {
				return fmt.Errorf("package manager cannot be empty")
			}
			if !c.IsValidPackageManager(lang, PackageManager(strVal)) {
				managers, _ := c.GetPackageManagersForLanguage(lang)
				return fmt.Errorf("invalid package manager for %s: %s. Supported managers: %v", lang, strVal, managers)
			}
		} else {
			return fmt.Errorf("package manager must be a string, got %T", value)
		}
	case InputFieldTestCommand, InputFieldBuildCommand, InputFieldLintCommand:
		if strVal, ok := value.(string); ok {
			if strVal == "" {
				return fmt.Errorf("%s command cannot be empty", inputField)
			}
		} else {
			return fmt.Errorf("%s command must be a string, got %T", inputField, value)
		}
	case InputFieldRequirements:
		if strVal, ok := value.(string); ok {
			if strVal == "" {
				return fmt.Errorf("requirements file path cannot be empty")
			}
		} else {
			return fmt.Errorf("requirements file path must be a string, got %T", value)
		}
	}
	return nil
}

// ValidateTemplateInputLegacy validates input values using legacy string-based field names
// Deprecated: Use ValidateTemplateInput with InputField instead
func (c *Configuration) ValidateTemplateInputLegacy(inputName string, value interface{}, lang Language) error {
	var inputField InputField
	switch inputName {
	case "nodeVersion":
		inputField = InputFieldNodeVersion
	case "goVersion":
		inputField = InputFieldGoVersion
	case "pythonVersion":
		inputField = InputFieldPythonVersion
	case "packageManager":
		inputField = InputFieldPackageManager
	case "testCommand":
		inputField = InputFieldTestCommand
	case "buildCommand":
		inputField = InputFieldBuildCommand
	case "lintCommand":
		inputField = InputFieldLintCommand
	case "requirements":
		inputField = InputFieldRequirements
	default:
		return fmt.Errorf("unknown input field: %s", inputName)
	}
	return c.ValidateTemplateInput(inputField, value, lang)
}

// ValidateAllInputs validates a complete set of template inputs for a language
func (c *Configuration) ValidateAllInputs(inputs map[InputField]interface{}, lang Language) error {
	if !c.IsValidLanguage(lang) {
		return fmt.Errorf("unsupported language: %s", lang)
	}

	var errors []string

	// Check for required fields based on language
	requiredFields := c.getRequiredFields(lang)
	for _, field := range requiredFields {
		if _, exists := inputs[field]; !exists {
			errors = append(errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	// Validate each provided input
	for field, value := range inputs {
		if err := c.ValidateTemplateInput(field, value, lang); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %v", errors)
	}

	return nil
}

// getRequiredFields returns the required input fields for a language
func (c *Configuration) getRequiredFields(lang Language) []InputField {
	switch lang {
	case LanguageGo:
		return []InputField{InputFieldGoVersion}
	case LanguageNode:
		return []InputField{InputFieldNodeVersion, InputFieldPackageManager}
	case LanguagePython:
		return []InputField{InputFieldPythonVersion, InputFieldPackageManager}
	default:
		return []InputField{}
	}
}

// ValidateAllInputsLegacy validates inputs using legacy string-based field names
// Deprecated: Use ValidateAllInputs with InputField keys instead
func (c *Configuration) ValidateAllInputsLegacy(inputs map[string]interface{}, lang Language) error {
	typedInputs := make(map[InputField]interface{})

	for strKey, value := range inputs {
		switch strKey {
		case "nodeVersion":
			typedInputs[InputFieldNodeVersion] = value
		case "goVersion":
			typedInputs[InputFieldGoVersion] = value
		case "pythonVersion":
			typedInputs[InputFieldPythonVersion] = value
		case "packageManager":
			typedInputs[InputFieldPackageManager] = value
		case "testCommand":
			typedInputs[InputFieldTestCommand] = value
		case "buildCommand":
			typedInputs[InputFieldBuildCommand] = value
		case "lintCommand":
			typedInputs[InputFieldLintCommand] = value
		case "requirements":
			typedInputs[InputFieldRequirements] = value
		default:
			return fmt.Errorf("unknown input field: %s", strKey)
		}
	}

	return c.ValidateAllInputs(typedInputs, lang)
}

// Type-safe version accessors

// GetGoVersions returns all supported Go versions
func (c *Configuration) GetGoVersions() []string {
	return c.Languages[LanguageGo].Versions
}

// GetNodeVersions returns all supported Node.js versions
func (c *Configuration) GetNodeVersions() []string {
	return c.Languages[LanguageNode].Versions
}

// GetPythonVersions returns all supported Python versions
func (c *Configuration) GetPythonVersions() []string {
	return c.Languages[LanguagePython].Versions
}

// GetVersionsForLanguage returns all supported versions for a given language
func (c *Configuration) GetVersionsForLanguage(lang Language) ([]string, error) {
	config, exists := c.Languages[lang]
	if !exists {
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
	return config.Versions, nil
}

// GetPackageManagersForLanguage returns all supported package managers for a given language
func (c *Configuration) GetPackageManagersForLanguage(lang Language) ([]PackageManager, error) {
	config, exists := c.Languages[lang]
	if !exists {
		return nil, fmt.Errorf("unsupported language: %s", lang)
	}
	return config.PackageManagers, nil
}

// IsValidLanguage checks if a language is supported
func (c *Configuration) IsValidLanguage(lang Language) bool {
	_, exists := c.Languages[lang]
	return exists
}

// IsValidInputFieldForLanguage checks if an input field is valid for a specific language
func (c *Configuration) IsValidInputFieldForLanguage(field InputField, lang Language) bool {
	validFields := c.GetValidInputFields(lang)
	for _, validField := range validFields {
		if validField == field {
			return true
		}
	}
	return false
}

// TypedDefaults provides type-safe access to default values (replaces legacy DefaultValues)
type TypedDefaults struct {
	config *Configuration
}

// NewTypedDefaults creates a new TypedDefaults instance
func NewTypedDefaults() *TypedDefaults {
	return &TypedDefaults{config: &Config}
}

// GetGoVersion returns the default Go version
func (td *TypedDefaults) GetGoVersion() string {
	return td.config.Languages[LanguageGo].DefaultVersion
}

// GetNodeVersion returns the default Node.js version
func (td *TypedDefaults) GetNodeVersion() string {
	return td.config.Languages[LanguageNode].DefaultVersion
}

// GetPythonVersion returns the default Python version
func (td *TypedDefaults) GetPythonVersion() string {
	return td.config.Languages[LanguagePython].DefaultVersion
}

// GetDefaultPackageManager returns the default package manager for a language
func (td *TypedDefaults) GetDefaultPackageManager(lang Language) (PackageManager, error) {
	if config, exists := td.config.Languages[lang]; exists {
		if config.DefaultManager != "" {
			return config.DefaultManager, nil
		}
		return "", fmt.Errorf("no default package manager for language %s", lang)
	}
	return "", fmt.Errorf("unsupported language: %s", lang)
}

// GetDefaultTestCommand returns the default test command for a language
func (td *TypedDefaults) GetDefaultTestCommand(lang Language) (string, error) {
	if config, exists := td.config.Languages[lang]; exists {
		return config.DefaultTestCmd, nil
	}
	return "", fmt.Errorf("unsupported language: %s", lang)
}

// GetDefaultBuildCommand returns the default build command for a language
func (td *TypedDefaults) GetDefaultBuildCommand(lang Language) (string, error) {
	if config, exists := td.config.Languages[lang]; exists {
		return config.DefaultBuildCmd, nil
	}
	return "", fmt.Errorf("unsupported language: %s", lang)
}

// GetDefaultLintCommand returns the default lint command for a language
func (td *TypedDefaults) GetDefaultLintCommand(lang Language) (string, error) {
	if config, exists := td.config.Languages[lang]; exists {
		return config.DefaultLintCmd, nil
	}
	return "", fmt.Errorf("unsupported language: %s", lang)
}

// GetDefaultRequirementsFile returns the default requirements file for a language
func (td *TypedDefaults) GetDefaultRequirementsFile(lang Language) (string, error) {
	if config, exists := td.config.Languages[lang]; exists {
		return config.DefaultReqFile, nil
	}
	return "", fmt.Errorf("unsupported language: %s", lang)
}

// GetSupportedLanguages returns all supported languages
func (td *TypedDefaults) GetSupportedLanguages() []Language {
	languages := make([]Language, 0, len(td.config.Languages))
	for lang := range td.config.Languages {
		languages = append(languages, lang)
	}
	return languages
}

// GetAllVersions returns all supported versions for all languages
func (td *TypedDefaults) GetAllVersions() map[Language][]string {
	versions := make(map[Language][]string)
	for lang, config := range td.config.Languages {
		versions[lang] = config.Versions
	}
	return versions
}

// GetAllPackageManagers returns all supported package managers for all languages
func (td *TypedDefaults) GetAllPackageManagers() map[Language][]PackageManager {
	managers := make(map[Language][]PackageManager)
	for lang, config := range td.config.Languages {
		if len(config.PackageManagers) > 0 {
			managers[lang] = config.PackageManagers
		}
	}
	return managers
}

// GetAllDefaults returns all default values for all languages
func (td *TypedDefaults) GetAllDefaults() map[Language]map[InputField]interface{} {
	allDefaults := make(map[Language]map[InputField]interface{})
	for lang := range td.config.Languages {
		allDefaults[lang] = td.config.GetAllDefaults(lang)
	}
	return allDefaults
}

// GetDefaultSecuritySeverity returns the default security severity level
func (td *TypedDefaults) GetDefaultSecuritySeverity() SecuritySeverity {
	return td.config.Security.DefaultLevel
}

// GetAllSecuritySeverities returns all supported security severity levels
func (td *TypedDefaults) GetAllSecuritySeverities() []SecuritySeverity {
	return td.config.Security.SeverityLevels
}

// Defaults is the global typed defaults instance (replaces legacy DefaultValues)
var Defaults = NewTypedDefaults()
