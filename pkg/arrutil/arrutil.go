package arrutil

// Strings type
type Strings []string

// Contains given element
func (ss Strings) Contains(sub string) bool {
	for _, s := range ss {
		if s == sub {
			return true
		}
	}
	return false
}

// Length element
func (ss Strings) Length() int {
	return len(ss)

}
