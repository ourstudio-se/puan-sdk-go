package puan

type Solution map[string]int

func (s Solution) Extract(variables ...string) Solution {
	extracted := make(Solution)
	for _, variable := range variables {
		extracted[variable] = s[variable]
	}

	return extracted
}
