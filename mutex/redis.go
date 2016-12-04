package mutex

type Redis interface {
	BRPopLPush(source string, destination string) (string, error)
	Bool(reply interface{}, err error) (bool, error)
	EvalScript(lua string, keys []string, args []string) (interface{}, error)
}
