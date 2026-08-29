package badgeurl

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
)

const (
	// DefaultBadge lets bestpractices.dev prompt the user to choose a target
	// baseline section after opening the generated edit URL.
	DefaultBadge             = "choose"
	defaultMaxURLLength      = 2000
	defaultJustificationSize = 240
)

var (
	// These are the BPB sections this utility knows how to target from
	// Privateer's current OSPS baseline mapping.
	supportedBadgeSections = map[string]struct{}{
		"choose":     {},
		"baseline-1": {},
		"baseline-2": {},
		"baseline-3": {},
	}
)

// Options configures badge URL generation.
//
// IncludeJustifications is a tri-state field: nil applies the package default
// of including justifications, while non-nil values explicitly enable or
// disable them.
type Options struct {
	Badge                 string
	IncludeJustifications *bool
}

type resultsFile struct {
	Payload struct {
		Config   *payloadConfig `yaml:"config"`
		RestData *struct {
			Config *payloadConfig `yaml:"config"`
		} `yaml:"restdata"`
	} `yaml:"payload"`
	EvaluationSuites []evaluationSuite `yaml:"evaluation-suites"`
}

type payloadConfig struct {
	Vars map[string]string `yaml:"vars"`
}

type evaluationSuite struct {
	ControlEvaluations controlEvaluations `yaml:"control-evaluations"`
}

type controlEvaluations struct {
	Evaluations []controlEvaluation `yaml:"evaluations"`
}

type controlEvaluation struct {
	AssessmentLogs []assessmentLog `yaml:"assessment-logs"`
}

type assessmentLog struct {
	Requirement struct {
		EntryID string `yaml:"entry-id"`
	} `yaml:"requirement"`
	Result        string   `yaml:"result"`
	Message       string   `yaml:"message"`
	Applicability []string `yaml:"applicability"`
}

type proposalUnit struct {
	key     string
	encoded string
}

// GenerateFromFile reads a serialized Privateer results file and returns one
// or more Best Practices Badge automation proposal URLs.
func GenerateFromFile(path string, options Options) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read results file: %w", err)
	}

	return Generate(content, options)
}

// Generate converts a serialized Privateer results document into one or more
// Best Practices Badge automation-proposals URLs.
// Reference: https://github.com/coreinfrastructure/best-practices-badge/blob/main/docs/automation-proposals.md
func Generate(content []byte, options Options) ([]string, error) {
	options = normalizeOptions(options)
	if err := validateOptions(options); err != nil {
		return nil, err
	}

	var results resultsFile
	if err := yaml.Unmarshal(content, &results); err != nil {
		return nil, fmt.Errorf("parse results YAML: %w", err)
	}

	repoURL, err := extractRepositoryURL(results)
	if err != nil {
		return nil, err
	}

	units := collectProposalUnits(results, options)
	if len(units) == 0 {
		return nil, fmt.Errorf("no supported Best Practices Badge links could be generated from the results")
	}

	baseURL := buildBaseURL(repoURL, options.Badge)
	return buildURLs(baseURL, units)
}

func normalizeOptions(options Options) Options {
	if strings.TrimSpace(options.Badge) == "" {
		options.Badge = DefaultBadge
	}
	if options.IncludeJustifications == nil {
		options.IncludeJustifications = boolPtr(true)
	}
	return options
}

func validateOptions(options Options) error {
	if _, ok := supportedBadgeSections[options.Badge]; !ok {
		return fmt.Errorf("invalid badge %q: must be one of choose, baseline-1, baseline-2, baseline-3", options.Badge)
	}
	return nil
}

// extractRepositoryURL finds the repository identity that BPB uses to look up
// the target project before applying proposal fields.
func extractRepositoryURL(results resultsFile) (string, error) {
	// Privateer results may carry config in more than one serialized payload
	// location depending on how the scanner wrote the output.
	configs := []*payloadConfig{
		results.Payload.Config,
	}
	if results.Payload.RestData != nil {
		configs = append(configs, results.Payload.RestData.Config)
	}

	for _, cfg := range configs {
		if cfg == nil {
			continue
		}
		owner := strings.TrimSpace(cfg.Vars["owner"])
		repo := strings.TrimSpace(cfg.Vars["repo"])
		if owner != "" && repo != "" {
			return fmt.Sprintf("https://github.com/%s/%s", owner, repo), nil
		}
	}

	return "", fmt.Errorf("could not determine repository URL from results payload")
}

// collectProposalUnits walks the evaluation logs, filters them to the target
// badge scope, and turns supported findings into stable query-string fragments.
func collectProposalUnits(results resultsFile, options Options) []proposalUnit {
	allowedLevels := levelsForBadge(options.Badge)
	seen := map[string]struct{}{}
	units := make([]proposalUnit, 0)

	for _, suite := range results.EvaluationSuites {
		for _, evaluation := range suite.ControlEvaluations.Evaluations {
			for _, log := range evaluation.AssessmentLogs {
				// Requirement IDs can appear more than once across suites. Keep the
				// first supported occurrence so the output is stable and non-duplicated.
				requirementID := strings.TrimSpace(log.Requirement.EntryID)
				if requirementID == "" {
					continue
				}
				if _, ok := seen[requirementID]; ok {
					continue
				}
				if !isApplicable(log.Applicability, allowedLevels) {
					continue
				}

				status, ok := mapResult(log.Result)
				if !ok {
					continue
				}

				key := badgeFieldName(requirementID)
				parts := []string{fmt.Sprintf("%s_status=%s", key, url.QueryEscape(status))}
				if *options.IncludeJustifications {
					justification := sanitizeJustification(log.Message)
					if justification != "" {
						parts = append(parts, fmt.Sprintf("%s_justification=%s", key, url.QueryEscape(justification)))
					}
				}

				units = append(units, proposalUnit{
					key:     key,
					encoded: strings.Join(parts, "&"),
				})
				seen[requirementID] = struct{}{}
			}
		}
	}

	sort.Slice(units, func(i, j int) bool {
		return units[i].key < units[j].key
	})

	return units
}

// levelsForBadge maps a BPB section to the Privateer OSPS maturity levels
// whose findings should be included in the generated link.
func levelsForBadge(badge string) map[string]struct{} {
	if badge == DefaultBadge {
		// "choose" defers section choice to BPB, so include all applicable levels.
		return nil
	}

	levels := map[string]struct{}{}
	for _, level := range []string{"maturity-1", "maturity-2", "maturity-3"} {
		levels[level] = struct{}{}
		if badge == "baseline-1" && level == "maturity-1" {
			break
		}
		if badge == "baseline-2" && level == "maturity-2" {
			break
		}
		if badge == "baseline-3" && level == "maturity-3" {
			break
		}
	}
	return levels
}

func isApplicable(applicability []string, allowedLevels map[string]struct{}) bool {
	if allowedLevels == nil || len(applicability) == 0 {
		return true
	}
	for _, level := range applicability {
		if _, ok := allowedLevels[level]; ok {
			return true
		}
	}
	return false
}

// mapResult converts Privateer's control result vocabulary into BPB's status
// vocabulary and drops unsupported states.
func mapResult(result string) (string, bool) {
	switch strings.TrimSpace(strings.ToLower(result)) {
	case "passed":
		return "Met", true
	case "failed":
		return "Unmet", true
	case "notapplicable", "not applicable", "n/a":
		return "N/A", true
	default:
		return "", false
	}
}

func badgeFieldName(requirementID string) string {
	replacer := strings.NewReplacer("-", "_", ".", "_")
	return strings.ToLower(replacer.Replace(requirementID))
}

// sanitizeJustification keeps reviewer context short and URL-safe so it can be
// embedded directly into a BPB proposal link.
func sanitizeJustification(message string) string {
	cleaned := strings.TrimSpace(message)
	if cleaned == "" {
		return ""
	}
	// Keep reviewer context compact while preserving evidence details such as
	// URLs; QueryEscape handles the actual URL encoding later.
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\r", " ")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	runes := []rune(cleaned)
	if len(runes) > defaultJustificationSize {
		cleaned = strings.TrimSpace(string(runes[:defaultJustificationSize]))
	}
	return cleaned
}

func boolPtr(value bool) *bool {
	return &value
}

func buildBaseURL(repoURL string, badge string) string {
	return fmt.Sprintf(
		"https://www.bestpractices.dev/projects?as=edit&section=%s&url=%s",
		url.QueryEscape(badge),
		url.QueryEscape(repoURL),
	)
}

// buildURLs emits one link when it fits within the default URL budget and
// otherwise batches proposal fragments into multiple links that can be applied
// in order.
func buildURLs(baseURL string, units []proposalUnit) ([]string, error) {
	urls := make([]string, 0, 1)
	current := baseURL
	hasUnits := false

	for _, unit := range units {
		candidate := current + "&" + unit.encoded
		if len(candidate) > defaultMaxURLLength {
			if !hasUnits {
				return nil, fmt.Errorf("a single Best Practices Badge proposal entry exceeds %d characters; disable justifications or shorten the source evidence text", defaultMaxURLLength)
			}
			urls = append(urls, current)
			current = baseURL + "&" + unit.encoded
			hasUnits = true
			continue
		}
		current = candidate
		hasUnits = true
	}

	if hasUnits {
		urls = append(urls, current)
	}

	return urls, nil
}
