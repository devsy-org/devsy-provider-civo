package main

import (
	"bufio"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	providerName = "civo"
	githubOwner  = "devsy-org"
	githubRepo   = "devsy-provider-civo"

	platformLinuxAMD64   = "linux/amd64"
	platformLinuxARM64   = "linux/arm64"
	platformDarwinAMD64  = "darwin/amd64"
	platformDarwinARM64  = "darwin/arm64"
	platformWindowsAMD64 = "windows/amd64"
)

type Provider struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description"`
	Icon         string            `yaml:"icon"`
	OptionGroups []OptionGroup     `yaml:"optionGroups"`
	Options      Options           `yaml:"options"`
	Agent        Agent             `yaml:"agent"`
	Binaries     Binaries          `yaml:"binaries"`
	Exec         map[string]string `yaml:"exec"`
}

type OptionGroup struct {
	Name           string   `yaml:"name"`
	DefaultVisible bool     `yaml:"defaultVisible"`
	Options        []string `yaml:"options"`
}

type Options map[string]Option

type Option struct {
	Description string   `yaml:"description,omitempty"`
	Required    bool     `yaml:"required,omitempty"`
	Password    bool     `yaml:"password,omitempty"`
	Default     string   `yaml:"default,omitempty"`
	Command     string   `yaml:"command,omitempty"`
	Suggestions []string `yaml:"suggestions,omitempty"`
	Local       bool     `yaml:"local,omitempty"`
	Hidden      bool     `yaml:"hidden,omitempty"`
	Cache       string   `yaml:"cache,omitempty"`
}

type Agent struct {
	Path                    string         `yaml:"path"`
	InactivityTimeout       string         `yaml:"inactivityTimeout"`
	InjectGitCredentials    string         `yaml:"injectGitCredentials"`
	InjectDockerCredentials string         `yaml:"injectDockerCredentials"`
	Binaries                map[string]any `yaml:"binaries"`
	Exec                    map[string]any `yaml:"exec"`
}

type Binaries struct {
	CivoProvider []Binary `yaml:"CIVO_PROVIDER"`
}

type Binary struct {
	OS       string `yaml:"os"`
	Arch     string `yaml:"arch"`
	Path     string `yaml:"path"`
	Checksum string `yaml:"checksum"`
}

type buildConfig struct {
	version     string
	projectRoot string
	isRelease   bool
	checksums   map[string]string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("expected version as argument")
	}

	cfg, err := newBuildConfig(os.Args[1])
	if err != nil {
		return err
	}

	provider, err := buildProvider(cfg)
	if err != nil {
		return err
	}

	output, err := yaml.Marshal(provider)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	_, err = os.Stdout.Write(output)
	return err
}

func newBuildConfig(version string) (*buildConfig, error) {
	checksums, err := parseChecksums("./dist/checksums.txt")
	if err != nil {
		return nil, fmt.Errorf("parse checksums: %w", err)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		owner := getEnvOrDefault("GITHUB_OWNER", githubOwner)
		projectRoot = fmt.Sprintf(
			"https://github.com/%s/%s/releases/download/%s",
			owner,
			githubRepo,
			version,
		)
	}

	isRelease := strings.Contains(projectRoot, "github.com") &&
		strings.Contains(projectRoot, "/releases/")

	return &buildConfig{
		version:     version,
		projectRoot: projectRoot,
		isRelease:   isRelease,
		checksums:   checksums,
	}, nil
}

func buildProvider(cfg *buildConfig) (Provider, error) {
	binaries, err := buildBinaries(cfg, allPlatforms())
	if err != nil {
		return Provider{}, err
	}
	agent, err := buildAgent(cfg)
	if err != nil {
		return Provider{}, err
	}
	return Provider{
		Name:         providerName,
		Version:      cfg.version,
		Description:  "Devsy on CIVO Cloud",
		Icon:         "https://raw.githubusercontent.com/devsy-org/devsy/main/desktop/src/images/civo.svg",
		OptionGroups: buildOptionGroups(),
		Options:      buildOptions(),
		Agent:        agent,
		Binaries:     binaries,
		Exec: map[string]string{
			"init":    "${CIVO_PROVIDER} init",
			"command": "${CIVO_PROVIDER} command",
			"create":  "${CIVO_PROVIDER} create",
			"delete":  "${CIVO_PROVIDER} delete",
			"start":   "${CIVO_PROVIDER} start",
			"stop":    "${CIVO_PROVIDER} stop",
			"status":  "${CIVO_PROVIDER} status",
		},
	}, nil
}

func buildOptionGroups() []OptionGroup {
	return []OptionGroup{
		{
			Name:           "CIVO options",
			DefaultVisible: true,
			Options: []string{
				"CIVO_DISK_SIZE",
				"CIVO_DISK_IMAGE",
				"CIVO_INSTANCE_TYPE",
			},
		},
		{
			Name:           "Agent options",
			DefaultVisible: false,
			Options: []string{
				"AGENT_PATH",
				"INACTIVITY_TIMEOUT",
				"INJECT_DOCKER_CREDENTIALS",
				"INJECT_GIT_CREDENTIALS",
			},
		},
	}
}

func buildOptions() Options {
	opts := Options{}
	maps.Copy(opts, buildCoreOptions())
	maps.Copy(opts, buildInstanceOptions())
	maps.Copy(opts, buildAgentOptions())
	opts["CIVO_TOKEN"] = Option{ //nolint:gosec // not actual credentials
		Local:       true,
		Hidden:      true,
		Cache:       "5m",
		Description: "The CIVO auth token to use.",
		Command:     "${CIVO_PROVIDER} token",
	}
	return opts
}

func buildCoreOptions() Options {
	return Options{
		"CIVO_API_KEY": {
			Description: "The civo api key to use.",
			Required:    true,
			Password:    true,
		},
		"CIVO_REGION": {
			Description: "The civo cloud region to create the VM in. E.g. LON1",
			Required:    true,
			Suggestions: []string{
				"FRA1",
				"LON1",
				"NYC1",
				"PHX1",
			},
		},
	}
}

func buildInstanceOptions() Options {
	return Options{
		"CIVO_DISK_SIZE": {
			Description: "The disk size to use.",
			Default:     "40",
		},
		"CIVO_DISK_IMAGE": {
			Description: "The disk image to use.",
			Default:     "d927ad2f-5073-4ed6-b2eb-b8e61aef29a8",
		},
		"CIVO_INSTANCE_TYPE": {
			Description: "The machine type to use.",
			Default:     "g3.large",
			Suggestions: []string{
				"g3.small",
				"g3.medium",
				"g3.large",
				"g3.xlarge",
				"g3.2xlarge",
			},
		},
	}
}

func buildAgentOptions() Options {
	return Options{
		"INACTIVITY_TIMEOUT": {
			Description: "If defined, will automatically stop the VM after the inactivity period.",
			Default:     "10m",
		},
		"INJECT_GIT_CREDENTIALS": {
			Description: "If Devsy should inject git credentials into the remote host.",
			Default:     "true",
		},
		"INJECT_DOCKER_CREDENTIALS": {
			Description: "If Devsy should inject docker credentials into the remote host.",
			Default:     "true",
		},
		"AGENT_PATH": {
			Description: "The path where to inject the Devsy agent to.",
			Default:     "/var/lib/toolbox/devsy",
		},
	}
}

//nolint:gosec // G101: template variables, not actual credentials
func buildAgent(cfg *buildConfig) (Agent, error) {
	linuxBins, err := buildBinaries(cfg, linuxPlatforms())
	if err != nil {
		return Agent{}, err
	}
	return Agent{
		Path:                    "${AGENT_PATH}",
		InactivityTimeout:       "${INACTIVITY_TIMEOUT}",
		InjectGitCredentials:    "${INJECT_GIT_CREDENTIALS}",
		InjectDockerCredentials: "${INJECT_DOCKER_CREDENTIALS}",
		Binaries: map[string]any{
			"CIVO_PROVIDER": linuxBins.CivoProvider,
		},
		Exec: map[string]any{
			"shutdown": "${CIVO_PROVIDER} stop",
		},
	}, nil
}

func buildBinaries(cfg *buildConfig, platforms []string) (Binaries, error) {
	list, err := buildBinaryList(cfg, platforms)
	if err != nil {
		return Binaries{}, err
	}
	return Binaries{CivoProvider: list}, nil
}

func buildBinaryList(cfg *buildConfig, platforms []string) ([]Binary, error) {
	result := make([]Binary, 0, len(platforms))
	for _, platform := range platforms {
		binary, err := buildBinary(cfg, platform)
		if err != nil {
			return nil, err
		}
		result = append(result, binary)
	}
	return result, nil
}

func buildBinary(cfg *buildConfig, platform string) (Binary, error) {
	os, arch, ok := strings.Cut(platform, "/")
	if !ok {
		return Binary{}, fmt.Errorf("invalid platform %q", platform)
	}

	path, err := buildBinaryPath(cfg, platform, os, arch)
	if err != nil {
		return Binary{}, err
	}

	filename := buildFilename(os, arch)
	checksum, ok := cfg.checksums[filename]
	if !ok || checksum == "" {
		return Binary{}, fmt.Errorf("missing checksum for %s", filename)
	}

	return Binary{
		OS:       os,
		Arch:     arch,
		Path:     path,
		Checksum: checksum,
	}, nil
}

func buildBinaryPath(cfg *buildConfig, platform, os, arch string) (string, error) {
	dir := buildDir(platform)
	if dir == "" {
		return "", fmt.Errorf("unsupported platform %q", platform)
	}

	basePath, err := resolveBasePath(cfg, dir)
	if err != nil {
		return "", err
	}

	filename := buildFilename(os, arch)
	return joinPath(basePath, filename)
}

func resolveBasePath(cfg *buildConfig, dir string) (string, error) {
	if cfg.isRelease {
		return cfg.projectRoot, nil
	}

	if strings.HasPrefix(cfg.projectRoot, "http://") ||
		strings.HasPrefix(cfg.projectRoot, "https://") {
		return joinURLPath(cfg.projectRoot, dir)
	}

	absPath, err := filepath.Abs(cfg.projectRoot)
	if err != nil {
		return "", fmt.Errorf("abs PROJECT_ROOT: %w", err)
	}
	return filepath.Join(absPath, dir), nil
}

func joinPath(basePath, filename string) (string, error) {
	if strings.HasPrefix(basePath, "http://") || strings.HasPrefix(basePath, "https://") {
		return joinURLPath(basePath, filename)
	}
	return filepath.Join(basePath, filename), nil
}

func joinURLPath(base, elem string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}
	joined, err := url.JoinPath(parsed.String(), elem)
	if err != nil {
		return "", fmt.Errorf("join URL path: %w", err)
	}
	return joined, nil
}

func buildFilename(os, arch string) string {
	filename := fmt.Sprintf("devsy-provider-%s-%s-%s", providerName, os, arch)
	if os == "windows" {
		filename += ".exe"
	}
	return filename
}

func buildDir(platform string) string {
	dirs := map[string]string{
		platformLinuxAMD64:   "build_linux_amd64_v1",
		platformLinuxARM64:   "build_linux_arm64_v8.0",
		platformDarwinAMD64:  "build_darwin_amd64_v1",
		platformDarwinARM64:  "build_darwin_arm64_v8.0",
		platformWindowsAMD64: "build_windows_amd64_v1",
	}
	return dirs[platform]
}

func allPlatforms() []string {
	return []string{
		platformLinuxAMD64,
		platformLinuxARM64,
		platformDarwinAMD64,
		platformDarwinARM64,
		platformWindowsAMD64,
	}
}

func linuxPlatforms() []string {
	return []string{platformLinuxAMD64, platformLinuxARM64}
}

func parseChecksums(path string) (map[string]string, error) {
	file, err := os.Open(path) //nolint:gosec // path is a build-time constant
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	checksums := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if checksum, filename, ok := strings.Cut(scanner.Text(), " "); ok {
			checksums[strings.TrimSpace(filename)] = checksum
		}
	}

	return checksums, scanner.Err()
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
