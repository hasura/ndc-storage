package encoding

const (
	ContentTypeTextPlain                 string = "text/plain"
	ContentTypeApplicationJSON           string = "application/json"
	contentTypeTextCSV                          = "text/csv"
	contentTypeTextXCSV                         = "text/x-csv"
	contentTypeTextCommaSeparatedValues         = "text/comma-separated-values"
	contentTypeTextXCommaSeparatedValues        = "text/x-comma-separated-values"
	contentTypeTextTabSeparatedValues           = "text/tab-separated-values"
	contentTypeApplicationCSV                   = "application/csv"
	contentTypeApplicationXCSV                  = "application/x-csv"
)

var enums_contentTypeCSV = []string{
	contentTypeTextCSV,
	contentTypeTextXCSV,
	contentTypeTextCommaSeparatedValues,
	contentTypeTextXCommaSeparatedValues,
	contentTypeTextTabSeparatedValues,
	contentTypeApplicationCSV,
	contentTypeApplicationXCSV,
}
