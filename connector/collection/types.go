package collection

// ColumnOrder the structured sorting columns.
type ColumnOrder struct {
	Name       string
	Descending bool
}

// StringComparisonOperator represents the explicit comparison expression for string columns.
type StringComparisonOperator struct {
	Value    string
	Operator string
}
