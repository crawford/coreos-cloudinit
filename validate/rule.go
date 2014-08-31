package validate

type Reporter interface {
	Error(line int, message string)
	Warning(line int, message string)
	Entries() []Entry
}

type ruleContext struct {
	content []byte
	currentLine int
}

type test struct {
	context ruleContext
	rule rule
}

type rule func(context ruleContext, validator *validator)

var (
	Rules []rule = YamlRules
)
