package index

// Profile stores index metrics for debugging and benchmarking.
type Profile struct {
	sortTime float64
}

// NewProfile creates a new Profile.
func NewProfile() *Profile {
	return &Profile{}
}

// SetSortingTime sets the sorting time.
func (p *Profile) SetSortingTime(time float64) {
	p.sortTime = time
}

// GetSortingTime returns the sorting time.
func (p *Profile) GetSortingTime() float64 {
	return p.sortTime
}
