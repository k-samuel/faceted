package intersection

// IntersectionInterface defines methods for computing intersections.
type IntersectionInterface interface {
	// GetIntersectMapCount returns the count of intersecting elements.
	// a: slice of record IDs
	// b: map of record IDs to check against
	GetIntersectMapCount(a []int, b map[int]bool) int

	// HasIntersectIntMap checks if two collections have any intersection.
	HasIntersectIntMap(a []int, b map[int]bool) bool
}
