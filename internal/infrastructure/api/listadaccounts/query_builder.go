package listadaccounts

import (
	"fmt"
	"regexp"
	"strings"
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
			return fmt.Errorf("listadaccounts: account ID %q is invalid", rawID)
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
