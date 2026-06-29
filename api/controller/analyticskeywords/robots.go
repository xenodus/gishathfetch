package analyticskeywords

import (
	"net/url"
	"sort"
	"strings"
)

const RobotsTxtObjectKey = "robots.txt"

// RobotsTxtBase is the static robots.txt body written before keyword URLs are appended.
const RobotsTxtBase = `User-agent: *
Allow: /

Disallow: /analytics/
`

// CollectUniqueSearchTerms returns deduplicated card search terms from every report period.
func CollectUniqueSearchTerms(report *Report) []string {
	if report == nil || len(report.Periods) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	terms := make([]string, 0)
	for _, period := range report.Periods {
		for _, keyword := range period.Keywords {
			term := strings.TrimSpace(keyword.Term)
			if term == "" {
				continue
			}
			if _, ok := seen[term]; ok {
				continue
			}
			seen[term] = struct{}{}
			terms = append(terms, term)
		}
	}

	sort.Strings(terms)
	return terms
}

// BuildSearchPagePath returns the canonical search page path for robots.txt Allow lines.
func BuildSearchPagePath(term string) string {
	query := url.Values{}
	query.Set("s", term)
	return "/?" + query.Encode()
}

// BuildSearchPageURL returns the canonical search page URL for a card name.
func BuildSearchPageURL(baseURL, term string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "https://gishathfetch.com/"
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return baseURL + strings.TrimPrefix(BuildSearchPagePath(term), "/")
}

// BuildRobotsTxt renders robots.txt with Allow entries for each unique search page URL.
func BuildRobotsTxt(report *Report, baseURL string) string {
	terms := CollectUniqueSearchTerms(report)
	if len(terms) == 0 {
		return RobotsTxtBase
	}

	var builder strings.Builder
	builder.WriteString(RobotsTxtBase)
	builder.WriteString("\n# Popular MTG card search pages (updated daily)\n")
	for _, term := range terms {
		builder.WriteString("Allow: ")
		builder.WriteString(BuildSearchPageURL(baseURL, term))
		builder.WriteByte('\n')
	}

	return builder.String()
}
