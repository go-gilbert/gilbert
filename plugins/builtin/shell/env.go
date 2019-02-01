package shell

// Environment is map of predefined scope variables
type Environment map[string]string

// Empty checks if predefined vars list is empty
func (e Environment) Empty() bool {
	return e == nil || (len(e) == 0)
}

func (e Environment) ToArray(defaults ...string) (arr []string) {
	arr = make([]string, len(defaults)+len(e))
	for k, v := range e {
		arr = append(arr, k+"="+v)
	}

	if len(defaults) == 0 {
		return arr
	}

	return append(arr, defaults...)
}
