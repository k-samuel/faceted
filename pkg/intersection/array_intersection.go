package intersection

// ArrayIntersection implements IntersectionInterface for standard arrays.
type ArrayIntersection struct{}

// NewArrayIntersection creates a new ArrayIntersection.
func NewArrayIntersection() *ArrayIntersection {
	return &ArrayIntersection{}
}

// GetIntersectMapCount returns the count of intersecting elements.
func (i *ArrayIntersection) GetIntersectMapCount(a []int, b map[int]bool) int {
	intersectLen := 0
	for _, key := range a {
		if b[key] {
			intersectLen++
		}
	}
	return intersectLen
}

// HasIntersectIntMap checks if two collections have any intersection.
func (i *ArrayIntersection) HasIntersectIntMap(a []int, b map[int]bool) bool {
	for _, key := range a {
		if b[key] {
			return true
		}
	}
	return false
}
