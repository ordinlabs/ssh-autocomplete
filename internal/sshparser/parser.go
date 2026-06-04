package sshparser

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// HostParser extracts Host entries from SSH config files.
type HostParser struct {
	CacheDir string
	CacheTTL time.Duration
}

// NewHostParser creates a HostParser with default settings.
func NewHostParser() *HostParser {
	return &HostParser{
		CacheDir: os.TempDir(),
		CacheTTL: 10 * time.Second,
	}
}

type cachedHosts struct {
	Timestamp time.Time `json:"timestamp"`
	Hosts     []string  `json:"hosts"`
}

func (p *HostParser) cacheFilePath() string {
	return filepath.Join(p.CacheDir, "ssh-autocomplete-cache.json")
}

// GetHosts returns SSH host names, using a cached result if available and fresh.
func (p *HostParser) GetHosts(useCache bool) ([]string, error) {
	if useCache {
		hosts, err := p.loadCache()
		if err == nil && hosts != nil {
			return hosts, nil
		}
	}

	sshConfigPath := defaultSSHConfigPath()
	hosts, err := p.ParseConfigFile(sshConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	hosts = deduplicateAndSort(hosts)
	_ = p.writeCache(hosts)

	return hosts, nil
}

// ParseConfigFile reads an SSH config file and extracts non-wildcard Host entries.
func (p *HostParser) ParseConfigFile(path string) ([]string, error) {
	visited := make(map[string]bool)
	return p.parseConfigFileRecursive(path, visited)
}

func (p *HostParser) parseConfigFileRecursive(path string, visited map[string]bool) ([]string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if visited[absPath] {
		return nil, nil
	}
	visited[absPath] = true

	file, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hosts []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if isHostDirective(line) {
			hostnames := extractHostnames(line)
			for _, h := range hostnames {
				if !isWildcard(h) {
					hosts = append(hosts, h)
				}
			}
			continue
		}

		if isIncludeDirective(line) {
			includedHosts, err := p.handleInclude(line, filepath.Dir(absPath), visited)
			if err != nil {
				continue
			}
			hosts = append(hosts, includedHosts...)
		}
	}

	if err := scanner.Err(); err != nil {
		return hosts, err
	}

	return hosts, nil
}

func (p *HostParser) handleInclude(line string, baseDir string, visited map[string]bool) ([]string, error) {
	pattern := extractIncludePattern(line)
	if pattern == "" {
		return nil, nil
	}

	if strings.HasPrefix(pattern, "~/") || strings.HasPrefix(pattern, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		pattern = filepath.Join(home, pattern[2:])
	} else if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(baseDir, pattern)
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var hosts []string
	for _, match := range matches {
		h, err := p.parseConfigFileRecursive(match, visited)
		if err != nil {
			continue
		}
		hosts = append(hosts, h...)
	}

	return hosts, nil
}

func (p *HostParser) loadCache() ([]string, error) {
	data, err := os.ReadFile(p.cacheFilePath())
	if err != nil {
		return nil, err
	}

	var cached cachedHosts
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}

	if time.Since(cached.Timestamp) > p.CacheTTL {
		return nil, nil
	}

	return cached.Hosts, nil
}

func (p *HostParser) writeCache(hosts []string) error {
	cached := cachedHosts{
		Timestamp: time.Now(),
		Hosts:     hosts,
	}

	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}

	return os.WriteFile(p.cacheFilePath(), data, 0600)
}

func defaultSSHConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "config")
}

func isHostDirective(line string) bool {
	lower := strings.ToLower(line)
	return strings.HasPrefix(lower, "host ") || strings.HasPrefix(lower, "host\t")
}

func isIncludeDirective(line string) bool {
	lower := strings.ToLower(line)
	return strings.HasPrefix(lower, "include ") || strings.HasPrefix(lower, "include\t")
}

func extractHostnames(line string) []string {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil
	}
	return parts[1:]
}

func extractIncludePattern(line string) string {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func isWildcard(hostname string) bool {
	return strings.ContainsAny(hostname, "*?")
}

func deduplicateAndSort(hosts []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, h := range hosts {
		if !seen[h] {
			seen[h] = true
			result = append(result, h)
		}
	}
	sort.Strings(result)
	return result
}
