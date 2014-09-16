package validate

type context struct {
	content []byte
	line    int
}

type test struct {
	context context
	rule    rule
}

type rule func(context context, validator *validator)

var (
	Rules []rule = YamlRules
)
