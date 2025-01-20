package manifest

// Vars is a set of declared variables
type Vars map[string]string

// Append appends variables from vars list
func (v Vars) Append(newVars Vars) (out Vars) {
	if v == nil {
		return newVars.Clone()
	}

	out = v.Clone()
	if newVars == nil || len(newVars) == 0 {
		return out
	}

	for k, val := range newVars {
		out[k] = val
	}

	return out
}

// AppendNew does the same as Append but doesn't overwrite existing values
func (v Vars) AppendNew(newVars Vars) (out Vars) {
	if v == nil {
		return newVars.Clone()
	}

	out = v.Clone()
	if newVars == nil || len(newVars) == 0 {
		return out
	}

	for k, val := range newVars {
		if _, ok := out[k]; ok {
			continue
		}
		out[k] = val
	}

	return out
}

// Clone creates a copy of variables map
func (v Vars) Clone() (out Vars) {
	out = make(Vars, len(v))
	for k, val := range v {
		out[k] = val
	}

	return out
}
