package notification

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

var _validator = validator.New()

func init() { //nolint:gochecknoinits
	endsWithAlNumRE := regexp.MustCompile(`\w$`)

	if err := _validator.RegisterValidation(
		"ends-with-alnum", MatchesPattern(endsWithAlNumRE)); err != nil {
		panic(err)
	}
}

func MatchesPattern(pattern *regexp.Regexp) validator.Func {
	return func(fl validator.FieldLevel) bool {
		return pattern.MatchString(fl.Field().String())
	}
}

// ConfigTree abstracts notification configs loaded from a filesystem.
type ConfigTree map[string]map[string]map[string]Config

// GetTeams returns a slice of teams from a ConfigTree.
func (t ConfigTree) GetTeams() []string {
	result := make([]string, 0, len(t))

	for team := range t {
		result = append(result, team)
	}

	return result
}

// GetAllNotifications returns a copy of the ConfigTree in native values.
func (t ConfigTree) GetAllNotifications() map[string]map[string]map[string]Config {
	result := make(map[string]map[string]map[string]Config)

	for team := range t {
		result[team] = GetTeamNotifications(team)
	}

	return result
}

// GetTeamNotifications returns a sub-tree for a particular team in native values.
func (t ConfigTree) GetTeamNotifications(team string) map[string]map[string]Config {
	teamTree, ok := t[team]
	if !ok {
		return nil
	}

	result := make(map[string]map[string]Config)

	for product := range teamTree {
		result[product] = t.GetProductNotifications(team, product)
	}

	return result
}

// GetProducts returns a slice of products for a particular team from a ConfigTree.
func (t ConfigTree) GetProducts(team string) []string {
	teamTree, ok := t[team]
	if !ok {
		return nil
	}

	result := make([]string, 0, len(t))

	for product := range teamTree {
		result = append(result, product)
	}

	return result
}

// GetProductNotifications returns a map from id -> Config for all notifications
// related to a team and product.
func (t ConfigTree) GetProductNotifications(team, product string) map[string]Config {
	teamTree, ok := t[team]
	if !ok {
		return nil
	}

	productTree, ok := teamTree[product]
	if !ok {
		return nil
	}

	result := make(map[string]Config)

	for id, cfg := range productTree {
		result[id] = cfg
	}

	return result
}

// GetNotification returns a Config and the value 'true' if a config exists
// for the given combination of 'team', 'product' and 'id'. Otherwise, an
// empty Config value is returned along with 'false'.
func (t ConfigTree) GetNotification(team, product, id string) (Config, bool) {
	teamTree, ok := t[team]
	if !ok {
		return Config{}, false
	}

	productTree, ok := teamTree[product]
	if !ok {
		return Config{}, false
	}

	cfg, ok := productTree[id]
	if !ok {
		return Config{}, false
	}

	return cfg, true
}

// GetAllNotifications returns all notifications loaded from the
// data directory within this package as a native map.
func GetAllNotifications() map[string]map[string]map[string]Config {
	return _notifications.GetAllNotifications()
}

// GetTeamNotifications returns all notifications loaded from the
// data directory within this package for a particular team as a
// native map.
func GetTeamNotifications(team string) map[string]map[string]Config {
	return _notifications.GetTeamNotifications(team)
}

// GetTeamNotifications returns all notifications loaded from the
// data directory within this package for a particular team and
// product as a native map.
func GetProductNotifications(team, product string) map[string]Config {
	return _notifications.GetProductNotifications(team, product)
}

// GetNotification returns a Config and the value 'true' if a config exists
// for the given combination of 'team', 'product' and 'id' from the
// data directory within this package. Otherwise, an empty Config
// value is returned along with 'false'.
func GetNotification(team, product, id string) (Config, bool) {
	return _notifications.GetNotification(team, product, id)
}

const _dataDirName = "data"

//go:embed data
var _dataDir embed.FS

var _notifications ConfigTree

func init() { //nolint:gochecknoinits
	var err error

	_notifications, err = loadNotifications(_dataDir)
	if err != nil {
		panic(fmt.Sprintf("unable to load customer notifcation configs: %v", err))
	}
}

func loadNotifications(dataDir fs.ReadDirFS) (ConfigTree, error) {
	teamDirs, err := getTeamDirectiories(dataDir)
	if err != nil {
		return nil, fmt.Errorf("retrieving team directories: %w", err)
	}

	result := make(ConfigTree)

	for _, dir := range teamDirs {
		teamName := dir.Name()

		subFS, err := fs.Sub(dataDir, path.Join(_dataDirName, teamName))
		if err != nil {
			return nil, fmt.Errorf("subbing team directory: %w", err)
		}

		teamNotifications, err := loadTeamConfigDirectory(subFS)
		if err != nil {
			return nil, fmt.Errorf("loading team config directory %q: %w", teamName, err)
		}

		result[teamName] = teamNotifications
	}

	return result, nil
}

func getTeamDirectiories(rootDir fs.ReadDirFS) ([]fs.DirEntry, error) {
	ents, err := rootDir.ReadDir(_dataDirName)
	if err != nil {
		return nil, fmt.Errorf("reading root directory entries: %w", err)
	}

	result := make([]fs.DirEntry, 0, len(ents))

	for _, ent := range ents {
		if !ent.IsDir() {
			continue
		}

		result = append(result, ent)
	}

	return result, nil
}

func loadTeamConfigDirectory(dir fs.FS) (map[string]map[string]Config, error) {
	ents, err := fs.ReadDir(dir, ".")
	if err != nil {
		return nil, fmt.Errorf("reading team directory: %w", err)
	}

	extPat := regexp.MustCompile(`^.*\.ya?ml$`)

	result := make(map[string]map[string]Config)

	for _, ent := range ents {
		name := ent.Name()

		if !extPat.MatchString(name) {
			continue
		}

		configFile, err := dir.Open(name)
		if err != nil {
			return nil, fmt.Errorf("opening config file %q: %w", name, err)
		}

		defer configFile.Close()

		configs, err := loadConfigFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("loading config file %q: %w", name, err)
		}

		key := strings.TrimSuffix(name, path.Ext(name))

		result[key] = configs
	}

	return result, nil
}

func loadConfigFile(f fs.File) (map[string]Config, error) {
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	result := make(map[string]Config)

	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshalling config yaml: %w", err)
	}

	for id, cfg := range result {
		if err := _validator.Struct(cfg); err != nil {
			return nil, fmt.Errorf("validating config %q: %w", id, err)
		}
	}

	return result, nil
}

// Config abstracts configuration values for a customer
// notification. Any changes to this struct should be
// updated in './data/README.md'.
type Config struct {
	Description  string `validate:"required,ends-with-alnum"`
	InternalOnly bool   `yaml:"internalOnly"`
	ServiceName  string `yaml:"serviceName"`
	Severity     string `validate:"required,oneof=Debug Error Fatal Info Warning"`
	Summary      string `validate:"required"`
}

const defaultServiceName = "SREManualAction"

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	type rawConfig Config

	raw := rawConfig{
		ServiceName: defaultServiceName,
	}

	if err := value.Decode(&raw); err != nil {
		return fmt.Errorf("unmarshalling raw notification config: %w", err)
	}

	*c = Config(raw)

	return nil
}

func (c *Config) ProvideRowData() map[string]interface{} {
	return map[string]interface{}{
		"Description":   c.Description,
		"Internal Only": c.InternalOnly,
		"Service Name":  c.ServiceName,
		"Severity":      c.Severity,
		"Summary":       c.Summary,
	}
}
