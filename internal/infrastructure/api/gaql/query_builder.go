package gaql

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	digitsOnlyRegex = regexp.MustCompile(`^\d+$`)
)

// QueryBuilder provides a fluent interface for building GAQL queries.
type QueryBuilder struct {
	resource string
	selects  []string
	wheres   []string
	orderBys []string
	limit    *int
}

// NewQueryBuilder creates a new QueryBuilder for the specified resource.
func NewQueryBuilder(resource string) *QueryBuilder {
	return &QueryBuilder{resource: resource}
}

// Select adds fields to the SELECT clause.
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.selects = append(qb.selects, fields...)
	return qb
}

// Where adds a condition to the WHERE clause.
func (qb *QueryBuilder) Where(cond string) *QueryBuilder {
	qb.wheres = append(qb.wheres, cond)
	return qb
}

// OrderBy adds fields to the ORDER BY clause.
func (qb *QueryBuilder) OrderBy(fields ...string) *QueryBuilder {
	qb.orderBys = append(qb.orderBys, fields...)
	return qb
}

// Limit sets the LIMIT clause.
func (qb *QueryBuilder) Limit(n int) *QueryBuilder {
	qb.limit = &n
	return qb
}

// Build constructs the final GAQL query string.
func (qb *QueryBuilder) Build() string {
	query := "SELECT " + strings.Join(qb.selects, ", ") +
		" FROM " + qb.resource
	if len(qb.wheres) > 0 {
		query += " WHERE " + strings.Join(qb.wheres, " AND ")
	}
	if len(qb.orderBys) > 0 {
		query += " ORDER BY " + strings.Join(qb.orderBys, ", ")
	}
	if qb.limit != nil {
		query += fmt.Sprintf(" LIMIT %d", *qb.limit)
	}
	return query
}

// WhereAccountIDs adds a WHERE clause to filter by account IDs
func (qb *QueryBuilder) WhereAccountIDs(ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(ids))
	for _, rawID := range ids {
		trimmed := strings.TrimSpace(rawID)
		if trimmed == "" {
			continue
		}

		var numericPart string
		if strings.HasPrefix(trimmed, "customers/") {
			// Extract the numeric part after "customers/"
			numericPart = strings.TrimPrefix(trimmed, "customers/")
		} else {
			// Use the entire string as the numeric part
			numericPart = trimmed
		}

		// Remove dashes and validate that the numeric part contains only digits
		sanitized := strings.ReplaceAll(numericPart, "-", "")
		if !digitsOnlyRegex.MatchString(sanitized) {
			return fmt.Errorf("account ID %q is invalid", rawID)
		}

		normalized = append(normalized, sanitized)
	}

	if len(normalized) == 0 {
		return nil
	}

	// Deduplicate IDs to avoid GAQL rejections.
	seen := make(map[string]bool)
	var unique []string
	for _, id := range normalized {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	// Convert IDs to resource name format, handling both cases:
	// - IDs without prefix: "9509030923" -> "customers/9509030923"
	// - IDs with prefix: "customers/9509030923" -> "customers/9509030923"
	var resourceNames []string
	for _, id := range unique {
		if strings.HasPrefix(id, "customers/") {
			// Already has the prefix, use as-is
			resourceNames = append(resourceNames, fmt.Sprintf("'%s'", id))
		} else {
			// Add the prefix
			resourceNames = append(resourceNames, fmt.Sprintf("'customers/%s'", id))
		}
	}
	qb.Where(fmt.Sprintf("customer_client.client_customer IN (%s)", strings.Join(resourceNames, ",")))
	return nil
}

// WhereAccountNames adds a WHERE clause to filter by account names
func (qb *QueryBuilder) WhereAccountNames(names []string) {
	if len(names) == 0 {
		return
	}

	var expressions []string
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}

		escaped := strings.ToLower(trimmed)
		escaped = strings.ReplaceAll(escaped, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "'", "''")
		escaped = strings.ReplaceAll(escaped, "%", "\\%")
		escaped = strings.ReplaceAll(escaped, "_", "\\_")

		expressions = append(expressions, fmt.Sprintf("customer_client.descriptive_name LIKE '%%%s%%'", escaped))
	}

	if len(expressions) == 0 {
		return
	}

	if len(expressions) == 1 {
		qb.Where(expressions[0])
	} else {
		qb.Where(strings.Join(expressions, " OR "))
	}
}

// WhereCampaignIDs adds a WHERE clause to filter by campaign IDs
// Campaign IDs should be numeric only (e.g., "123456789")
func (qb *QueryBuilder) WhereCampaignIDs(ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(ids))
	for _, rawID := range ids {
		trimmed := strings.TrimSpace(rawID)
		if trimmed == "" {
			continue
		}

		// Extract numeric part from resource name format if present
		// e.g., "customers/123456789/campaigns/987654321" -> "987654321"
		var numericID string
		if strings.Contains(trimmed, "/campaigns/") {
			parts := strings.Split(trimmed, "/campaigns/")
			if len(parts) == 2 {
				numericID = strings.TrimSpace(parts[1])
			} else {
				return fmt.Errorf("campaign ID %q is invalid: expected format 'customers/XXX/campaigns/YYY' or numeric ID", rawID)
			}
		} else {
			numericID = trimmed
		}

		// Remove dashes and validate that the ID contains only digits
		sanitized := strings.ReplaceAll(numericID, "-", "")
		if !digitsOnlyRegex.MatchString(sanitized) {
			return fmt.Errorf("campaign ID %q is invalid: must be numeric", rawID)
		}

		normalized = append(normalized, sanitized)
	}

	if len(normalized) == 0 {
		return nil
	}

	// Deduplicate IDs to avoid GAQL rejections.
	seen := make(map[string]bool)
	var unique []string
	for _, id := range normalized {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	// Build IN clause with campaign IDs
	if len(unique) == 1 {
		qb.Where(fmt.Sprintf("campaign.id = %s", unique[0]))
	} else {
		qb.Where(fmt.Sprintf("campaign.id IN (%s)", strings.Join(unique, ",")))
	}
	return nil
}

// WhereCampaignNames adds a WHERE clause to filter by campaign names
func (qb *QueryBuilder) WhereCampaignNames(names []string) {
	if len(names) == 0 {
		return
	}

	var expressions []string
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}

		escaped := strings.ToLower(trimmed)
		escaped = strings.ReplaceAll(escaped, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "'", "''")
		escaped = strings.ReplaceAll(escaped, "%", "\\%")
		escaped = strings.ReplaceAll(escaped, "_", "\\_")

		expressions = append(expressions, fmt.Sprintf("campaign.name LIKE '%%%s%%'", escaped))
	}

	if len(expressions) == 0 {
		return
	}

	if len(expressions) == 1 {
		qb.Where(expressions[0])
	} else {
		qb.Where(strings.Join(expressions, " OR "))
	}
}

// WhereStatus adds a WHERE clause to filter by campaign status
// Valid statuses: ENABLED, PAUSED, REMOVED
func (qb *QueryBuilder) WhereStatus(statuses []string) error {
	if len(statuses) == 0 {
		return nil
	}

	validStatuses := map[string]bool{
		"ENABLED": true,
		"PAUSED":  true,
		"REMOVED": true,
	}

	normalized := make([]string, 0, len(statuses))
	for _, status := range statuses {
		trimmed := strings.TrimSpace(strings.ToUpper(status))
		if trimmed == "" {
			continue
		}

		if !validStatuses[trimmed] {
			return fmt.Errorf("invalid status %q: must be one of ENABLED, PAUSED, REMOVED", status)
		}

		normalized = append(normalized, trimmed)
	}

	if len(normalized) == 0 {
		return nil
	}

	// Deduplicate statuses
	seen := make(map[string]bool)
	var unique []string
	for _, status := range normalized {
		if !seen[status] {
			seen[status] = true
			unique = append(unique, status)
		}
	}

	if len(unique) == 1 {
		qb.Where(fmt.Sprintf("campaign.status = %s", unique[0]))
	} else {
		qb.Where(fmt.Sprintf("campaign.status IN (%s)", strings.Join(unique, ",")))
	}
	return nil
}

// WhereDateRange adds a WHERE clause to filter by date range
// start and end should be in YYYY-MM-DD format
func (qb *QueryBuilder) WhereDateRange(start, end string) error {
	if start == "" && end == "" {
		return nil
	}

	// Validate and parse start date
	if start != "" {
		_, err := time.Parse("2006-01-02", start)
		if err != nil {
			return fmt.Errorf("invalid date format for start date %q: expected YYYY-MM-DD", start)
		}
	}

	// Validate and parse end date
	if end != "" {
		_, err := time.Parse("2006-01-02", end)
		if err != nil {
			return fmt.Errorf("invalid date format for end date %q: expected YYYY-MM-DD", end)
		}
	}

	if start != "" && end != "" {
		// Both dates provided - use BETWEEN
		qb.Where(fmt.Sprintf("segments.date BETWEEN '%s' AND '%s'", start, end))
	} else if start != "" {
		// Only start date - use >=
		qb.Where(fmt.Sprintf("segments.date >= '%s'", start))
	} else if end != "" {
		// Only end date - use <=
		qb.Where(fmt.Sprintf("segments.date <= '%s'", end))
	}

	return nil
}

