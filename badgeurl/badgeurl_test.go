package badgeurl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSingleURL(t *testing.T) {
	filePath := writeResultsFile(t, minimalResultsYAML())

	urls, err := GenerateFromFile(filePath, Options{Badge: "baseline-1", IncludeJustifications: boolOption(true)})
	if err != nil {
		t.Fatalf("GenerateFromFile returned error: %v", err)
	}
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}

	url := urls[0]
	if !strings.Contains(url, "section=baseline-1") {
		t.Fatalf("expected baseline section in URL: %s", url)
	}
	if !strings.Contains(url, "url=https%3A%2F%2Fgithub.com%2Facme%2Frocket") {
		t.Fatalf("expected repository URL in URL: %s", url)
	}
	if !strings.Contains(url, "osps_ac_01_01_status=Met") {
		t.Fatalf("expected mapped status in URL: %s", url)
	}
	if !strings.Contains(url, "osps_ac_01_01_justification=MFA+required") {
		t.Fatalf("expected justification in URL: %s", url)
	}
	if strings.Contains(url, "osps_do_03_01") {
		t.Fatalf("did not expect level 2 criterion in baseline-1 URL: %s", url)
	}
	if strings.Contains(url, "osps_do_04_01") {
		t.Fatalf("did not expect NeedsReview criterion in URL: %s", url)
	}
}

func TestGenerateMultipleURLsAboveDefaultLengthLimit(t *testing.T) {
	const proposalCount = 28
	filePath := writeResultsFile(t, longResultsYAML(28))

	urls, err := GenerateFromFile(filePath, Options{Badge: DefaultBadge, IncludeJustifications: boolOption(true)})
	if err != nil {
		t.Fatalf("GenerateFromFile returned error: %v", err)
	}
	if len(urls) < 2 {
		t.Fatalf("expected multiple URLs above the default length limit, got %d", len(urls))
	}
	for _, generatedURL := range urls {
		if len(generatedURL) > 2000 {
			t.Fatalf("expected URL length <= %d, got %d", 2000, len(generatedURL))
		}
	}

	joined := strings.Join(urls, "\n")
	if got := strings.Count(joined, "_status="); got != proposalCount {
		t.Fatalf("expected %d statuses across all split URLs, got %d", proposalCount, got)
	}
	if got := strings.Count(joined, "_justification="); got != proposalCount {
		t.Fatalf("expected %d justifications across all split URLs, got %d", proposalCount, got)
	}
	for i := 1; i <= proposalCount; i++ {
		statusKey := fmt.Sprintf("osps_ac_%02d_01_status=Met", i)
		if got := strings.Count(joined, statusKey); got != 1 {
			t.Fatalf("expected %q exactly once across all split URLs, got %d", statusKey, got)
		}
	}
}

func TestRejectsInvalidBadge(t *testing.T) {
	_, err := Generate([]byte(minimalResultsYAML()), Options{Badge: "gold", IncludeJustifications: boolOption(true)})
	if err == nil {
		t.Fatal("expected an error for invalid badge")
	}
}

func TestMissingInputFile(t *testing.T) {
	_, err := GenerateFromFile(filepath.Join(t.TempDir(), "missing.yaml"), Options{Badge: DefaultBadge, IncludeJustifications: boolOption(true)})
	if err == nil {
		t.Fatal("expected an error for missing file")
	}
}

func TestMalformedYAML(t *testing.T) {
	filePath := writeResultsFile(t, "payload: [\n")

	_, err := GenerateFromFile(filePath, Options{Badge: DefaultBadge, IncludeJustifications: boolOption(true)})
	if err == nil {
		t.Fatal("expected an error for malformed YAML")
	}
}

func TestNoSupportedLinks(t *testing.T) {
	filePath := writeResultsFile(t, unsupportedResultsYAML())

	_, err := GenerateFromFile(filePath, Options{Badge: DefaultBadge, IncludeJustifications: boolOption(true)})
	if err == nil {
		t.Fatal("expected an error when there are no supported links")
	}
}

func TestOmitJustificationsWhenDisabled(t *testing.T) {
	filePath := writeResultsFile(t, minimalResultsYAML())

	urls, err := GenerateFromFile(filePath, Options{Badge: DefaultBadge, IncludeJustifications: boolOption(false)})
	if err != nil {
		t.Fatalf("GenerateFromFile returned error: %v", err)
	}
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}
	if strings.Contains(urls[0], "_justification=") {
		t.Fatalf("did not expect justification in URL: %s", urls[0])
	}
}

func TestIncludeJustificationsByDefault(t *testing.T) {
	filePath := writeResultsFile(t, minimalResultsYAML())

	urls, err := GenerateFromFile(filePath, Options{Badge: DefaultBadge})
	if err != nil {
		t.Fatalf("GenerateFromFile returned error: %v", err)
	}
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}
	if !strings.Contains(urls[0], "_justification=") {
		t.Fatalf("expected justification in URL by default: %s", urls[0])
	}
}

func TestMapNotApplicableToNA(t *testing.T) {
	filePath := writeResultsFile(t, notApplicableResultsYAML())

	urls, err := GenerateFromFile(filePath, Options{Badge: DefaultBadge, IncludeJustifications: boolOption(true)})
	if err != nil {
		t.Fatalf("GenerateFromFile returned error: %v", err)
	}
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}
	if !strings.Contains(urls[0], "osps_le_01_01_status=N%2FA") {
		t.Fatalf("expected NotApplicable to map to N/A: %s", urls[0])
	}
}

func TestPreserveEvidenceURLInJustification(t *testing.T) {
	filePath := writeResultsFile(t, urlJustificationResultsYAML())

	urls, err := GenerateFromFile(filePath, Options{Badge: DefaultBadge, IncludeJustifications: boolOption(true)})
	if err != nil {
		t.Fatalf("GenerateFromFile returned error: %v", err)
	}
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}
	if !strings.Contains(urls[0], "https%3A%2F%2Fexample.com%2Fevidence%3Fcheck%3Dbranch-protection%26source%3Dscanner") {
		t.Fatalf("expected evidence URL to be preserved in justification: %s", urls[0])
	}
}

func TestTruncateJustificationOnRuneBoundary(t *testing.T) {
	filePath := writeResultsFile(t, unicodeJustificationResultsYAML())

	urls, err := GenerateFromFile(filePath, Options{Badge: DefaultBadge, IncludeJustifications: boolOption(true)})
	if err != nil {
		t.Fatalf("GenerateFromFile returned error: %v", err)
	}
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}
	if strings.Contains(urls[0], "%EF%BF%BD") {
		t.Fatalf("expected justification truncation to preserve UTF-8 rune boundaries: %s", urls[0])
	}
	if !strings.Contains(urls[0], "%C3%A9") {
		t.Fatalf("expected justification to preserve unicode characters: %s", urls[0])
	}
	if strings.Contains(urls[0], strings.Repeat("a", 260)) {
		t.Fatalf("expected justification to be truncated: %s", urls[0])
	}
}

func TestGenerateFromSerializedPayloadConfigShape(t *testing.T) {
	filePath := writeResultsFile(t, serializedPayloadResultsYAML())

	urls, err := GenerateFromFile(filePath, Options{Badge: DefaultBadge, IncludeJustifications: boolOption(true)})
	if err != nil {
		t.Fatalf("GenerateFromFile returned error: %v", err)
	}
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(urls))
	}

	url := urls[0]
	if !strings.Contains(url, "url=https%3A%2F%2Fgithub.com%2Fradius-project%2Fradius") {
		t.Fatalf("expected repository URL from serialized payload config: %s", url)
	}
	if !strings.Contains(url, "osps_ac_01_01_status=Met") {
		t.Fatalf("expected suffixed status field in URL: %s", url)
	}
	if !strings.Contains(url, "osps_ac_01_01_justification=This+control+is+enforced+by+GitHub") {
		t.Fatalf("expected justification in URL: %s", url)
	}
}

func writeResultsFile(t *testing.T, content string) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), "results.yaml")
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("write results file: %v", err)
	}
	return filePath
}

func boolOption(value bool) *bool {
	return &value
}

func minimalResultsYAML() string {
	return `payload:
  config:
    vars:
      owner: acme
      repo: rocket
evaluation-suites:
  - control-evaluations:
      evaluations:
        - assessment-logs:
            - requirement:
                entry-id: OSPS-AC-01.01
              result: Passed
              message: MFA required
              applicability:
                - maturity-1
                - maturity-2
            - requirement:
                entry-id: OSPS-DO-03.01
              result: Failed
              message: Missing provenance
              applicability:
                - maturity-2
            - requirement:
                entry-id: OSPS-DO-04.01
              result: NeedsReview
              message: Requires manual review
              applicability:
                - maturity-1
`
}

func unsupportedResultsYAML() string {
	return `payload:
  config:
    vars:
      owner: acme
      repo: rocket
evaluation-suites:
  - control-evaluations:
      evaluations:
        - assessment-logs:
            - requirement:
                entry-id: OSPS-DO-04.01
              result: NeedsReview
              message: Requires manual review
              applicability:
                - maturity-1
`
}

func notApplicableResultsYAML() string {
	return "payload:\n" +
		"  config:\n" +
		"    vars:\n" +
		"      owner: acme\n" +
		"      repo: rocket\n" +
		"evaluation-suites:\n" +
		"  - control-evaluations:\n" +
		"      evaluations:\n" +
		"        - assessment-logs:\n" +
		"            - requirement:\n" +
		"                entry-id: OSPS-LE-01.01\n" +
		"              result: NotApplicable\n" +
		"              message: Repository has no releases\n" +
		"              applicability:\n" +
		"                - maturity-1\n"
}

func urlJustificationResultsYAML() string {
	return "payload:\n" +
		"  config:\n" +
		"    vars:\n" +
		"      owner: acme\n" +
		"      repo: rocket\n" +
		"evaluation-suites:\n" +
		"  - control-evaluations:\n" +
		"      evaluations:\n" +
		"        - assessment-logs:\n" +
		"            - requirement:\n" +
		"                entry-id: OSPS-AC-01.01\n" +
		"              result: Passed\n" +
		"              message: \"Evidence: https://example.com/evidence?check=branch-protection&source=scanner\"\n" +
		"              applicability:\n" +
		"                - maturity-1\n"
}

func unicodeJustificationResultsYAML() string {
	return "payload:\n" +
		"  config:\n" +
		"    vars:\n" +
		"      owner: acme\n" +
		"      repo: rocket\n" +
		"evaluation-suites:\n" +
		"  - control-evaluations:\n" +
		"      evaluations:\n" +
		"        - assessment-logs:\n" +
		"            - requirement:\n" +
		"                entry-id: OSPS-AC-01.01\n" +
		"              result: Passed\n" +
		fmt.Sprintf("              message: \"%s\"\n", strings.Repeat("a", 239)+"é"+strings.Repeat("b", 40)) +
		"              applicability:\n" +
		"                - maturity-1\n"
}

func serializedPayloadResultsYAML() string {
	return "service-name: my-scan\n" +
		"plugin-name: github-repo\n" +
		"payload:\n" +
		"  restdata:\n" +
		"    config:\n" +
		"      vars:\n" +
		"        owner: radius-project\n" +
		"        repo: radius\n" +
		"  config:\n" +
		"    vars:\n" +
		"      owner: radius-project\n" +
		"      repo: radius\n" +
		"evaluation-suites:\n" +
		"  - control-evaluations:\n" +
		"      evaluations:\n" +
		"        - assessment-logs:\n" +
		"            - requirement:\n" +
		"                entry-id: OSPS-AC-01.01\n" +
		"              result: Passed\n" +
		"              message: This control is enforced by GitHub\n" +
		"              applicability:\n" +
		"                - maturity-1\n"
}

func longResultsYAML(count int) string {
	var builder strings.Builder
	builder.WriteString("payload:\n")
	builder.WriteString("  config:\n")
	builder.WriteString("    vars:\n")
	builder.WriteString("      owner: acme\n")
	builder.WriteString("      repo: rocket\n")
	builder.WriteString("evaluation-suites:\n")
	builder.WriteString("  - control-evaluations:\n")
	builder.WriteString("      evaluations:\n")
	builder.WriteString("        - assessment-logs:\n")
	for i := 1; i <= count; i++ {
		_, _ = fmt.Fprintf(&builder, "            - requirement:\n                entry-id: OSPS-AC-%02d.01\n", i)
		builder.WriteString("              result: Passed\n")
		_, _ = fmt.Fprintf(&builder, "              message: %s\n", strings.Repeat("evidence ", 24))
		builder.WriteString("              applicability:\n")
		builder.WriteString("                - maturity-1\n")
	}
	return builder.String()
}
