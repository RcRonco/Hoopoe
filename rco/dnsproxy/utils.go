package dnsproxy

const (
	DNS_REGEXPR = `^(([\{\}a-zA-Z0-9]|[\{\}a-zA-Z0-9][\{\}a-zA-Z0-9\-]*[\{\}a-zA-Z0-9])\.)*([\{\}A-Za-z0-9]|[\{\}A-Za-z0-9][\{\}A-Za-z0-9\-]*[\{\}A-Za-z0-9])$`
)

func ValidateTemplateBrackets(pattern string) bool {
	openBr := false
	// Run on each char
	for c := range pattern {
		// If char is open bracket mark open bracket state
		if c == '{' {
			// If there is already open bracket return invalid
			if openBr {
				return false
			}
			openBr = true
		}
		// If char is closing bracket mark closed bracket state
		if c == '}' {
			// If there is closing bracket without opened bracket return invalid
			if !openBr {
				return false
			}
			openBr = false
		}
	}
	return !openBr
}
