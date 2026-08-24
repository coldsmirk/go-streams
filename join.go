package streams

// --- Join Operations for Stream2 ---

// JoinResult holds the result of a join operation.
type JoinResult[K, V1, V2 any] struct {
	Key   K
	Left  V1
	Right V2
}

// JoinResultOptional holds the result of an outer join operation.
type JoinResultOptional[K, V1, V2 any] struct {
	Key   K
	Left  Optional[V1]
	Right Optional[V2]
}

// InnerJoin performs an inner join between two Stream2s on their keys.
// Only pairs with matching keys in both streams are included.
// Note: The second stream is collected into memory for the join.
func InnerJoin[K comparable, V1, V2 any](s1 Stream2[K, V1], s2 Stream2[K, V2]) Stream[JoinResult[K, V1, V2]] {
	return Stream[JoinResult[K, V1, V2]]{
		seq: func(yield func(JoinResult[K, V1, V2]) bool) {
			// Build lookup map from s2
			lookup := make(map[K][]V2)
			for k, v := range s2.seq {
				lookup[k] = append(lookup[k], v)
			}

			// Join with s1
			for k, v1 := range s1.seq {
				if v2s, ok := lookup[k]; ok {
					for _, v2 := range v2s {
						if !yield(JoinResult[K, V1, V2]{Key: k, Left: v1, Right: v2}) {
							return
						}
					}
				}
			}
		},
	}
}

// LeftJoin performs a left outer join between two Stream2s.
// All pairs from the left stream are included; right values are None if no match.
// Note: The second stream is collected into memory for the join.
func LeftJoin[K comparable, V1, V2 any](s1 Stream2[K, V1], s2 Stream2[K, V2]) Stream[JoinResultOptional[K, V1, V2]] {
	return Stream[JoinResultOptional[K, V1, V2]]{
		seq: func(yield func(JoinResultOptional[K, V1, V2]) bool) {
			// Build lookup map from s2
			lookup := make(map[K][]V2)
			for k, v := range s2.seq {
				lookup[k] = append(lookup[k], v)
			}

			// Left join with s1
			for k, v1 := range s1.seq {
				if v2s, ok := lookup[k]; ok {
					for _, v2 := range v2s {
						if !yield(JoinResultOptional[K, V1, V2]{
							Key:   k,
							Left:  Some(v1),
							Right: Some(v2),
						}) {
							return
						}
					}
				} else {
					if !yield(JoinResultOptional[K, V1, V2]{
						Key:   k,
						Left:  Some(v1),
						Right: None[V2](),
					}) {
						return
					}
				}
			}
		},
	}
}

// RightJoin performs a right outer join between two Stream2s.
// All pairs from the right stream are included; left values are None if no match.
// Note: Both streams are collected into memory for the join.
func RightJoin[K comparable, V1, V2 any](s1 Stream2[K, V1], s2 Stream2[K, V2]) Stream[JoinResultOptional[K, V1, V2]] {
	return Stream[JoinResultOptional[K, V1, V2]]{
		seq: func(yield func(JoinResultOptional[K, V1, V2]) bool) {
			// Build lookup map from s1
			lookup := make(map[K][]V1)
			for k, v := range s1.seq {
				lookup[k] = append(lookup[k], v)
			}

			// Right join with s2
			for k, v2 := range s2.seq {
				if v1s, ok := lookup[k]; ok {
					for _, v1 := range v1s {
						if !yield(JoinResultOptional[K, V1, V2]{
							Key:   k,
							Left:  Some(v1),
							Right: Some(v2),
						}) {
							return
						}
					}
				} else {
					if !yield(JoinResultOptional[K, V1, V2]{
						Key:   k,
						Left:  None[V1](),
						Right: Some(v2),
					}) {
						return
					}
				}
			}
		},
	}
}

// FullJoin performs a full outer join between two Stream2s.
// All pairs from both streams are included; missing values are None.
// Note: Both streams are collected into memory for the join.
func FullJoin[K comparable, V1, V2 any](s1 Stream2[K, V1], s2 Stream2[K, V2]) Stream[JoinResultOptional[K, V1, V2]] {
	return Stream[JoinResultOptional[K, V1, V2]]{
		seq: func(yield func(JoinResultOptional[K, V1, V2]) bool) {
			// Collect both streams
			left := make(map[K][]V1)
			for k, v := range s1.seq {
				left[k] = append(left[k], v)
			}

			right := make(map[K][]V2)
			for k, v := range s2.seq {
				right[k] = append(right[k], v)
			}

			// Track which right keys have been matched
			matchedRight := make(map[K]bool)

			// Iterate left, matching with right
			for k, v1s := range left {
				if v2s, ok := right[k]; ok {
					matchedRight[k] = true
					for _, v1 := range v1s {
						for _, v2 := range v2s {
							if !yield(JoinResultOptional[K, V1, V2]{
								Key:   k,
								Left:  Some(v1),
								Right: Some(v2),
							}) {
								return
							}
						}
					}
				} else {
					for _, v1 := range v1s {
						if !yield(JoinResultOptional[K, V1, V2]{
							Key:   k,
							Left:  Some(v1),
							Right: None[V2](),
						}) {
							return
						}
					}
				}
			}

			// Add unmatched right entries
			for k, v2s := range right {
				if !matchedRight[k] {
					for _, v2 := range v2s {
						if !yield(JoinResultOptional[K, V1, V2]{
							Key:   k,
							Left:  None[V1](),
							Right: Some(v2),
						}) {
							return
						}
					}
				}
			}
		},
	}
}

// --- Simplified Join with Default Values ---

// LeftJoinWith performs a left join with a default value for missing right values.
func LeftJoinWith[K comparable, V1, V2 any](s1 Stream2[K, V1], s2 Stream2[K, V2], defaultV2 V2) Stream[JoinResult[K, V1, V2]] {
	return Stream[JoinResult[K, V1, V2]]{
		seq: func(yield func(JoinResult[K, V1, V2]) bool) {
			lookup := make(map[K][]V2)
			for k, v := range s2.seq {
				lookup[k] = append(lookup[k], v)
			}

			for k, v1 := range s1.seq {
				if v2s, ok := lookup[k]; ok {
					for _, v2 := range v2s {
						if !yield(JoinResult[K, V1, V2]{Key: k, Left: v1, Right: v2}) {
							return
						}
					}
				} else {
					if !yield(JoinResult[K, V1, V2]{Key: k, Left: v1, Right: defaultV2}) {
						return
					}
				}
			}
		},
	}
}

// RightJoinWith performs a right join with a default value for missing left values.
func RightJoinWith[K comparable, V1, V2 any](s1 Stream2[K, V1], s2 Stream2[K, V2], defaultV1 V1) Stream[JoinResult[K, V1, V2]] {
	return Stream[JoinResult[K, V1, V2]]{
		seq: func(yield func(JoinResult[K, V1, V2]) bool) {
			lookup := make(map[K][]V1)
			for k, v := range s1.seq {
				lookup[k] = append(lookup[k], v)
			}

			for k, v2 := range s2.seq {
				if v1s, ok := lookup[k]; ok {
					for _, v1 := range v1s {
						if !yield(JoinResult[K, V1, V2]{Key: k, Left: v1, Right: v2}) {
							return
						}
					}
				} else {
					if !yield(JoinResult[K, V1, V2]{Key: k, Left: defaultV1, Right: v2}) {
						return
					}
				}
			}
		},
	}
}

// --- CoGroup Operations ---

// CoGrouped holds grouped values from two streams with the same key.
type CoGrouped[K, V1, V2 any] struct {
	Key   K
	Left  []V1
	Right []V2
}

// CoGroup groups values from two streams by their keys.
// Similar to SQL's FULL OUTER JOIN but groups all matching values together.
func CoGroup[K comparable, V1, V2 any](s1 Stream2[K, V1], s2 Stream2[K, V2]) Stream[CoGrouped[K, V1, V2]] {
	return Stream[CoGrouped[K, V1, V2]]{
		seq: func(yield func(CoGrouped[K, V1, V2]) bool) {
			// Collect both streams
			left := make(map[K][]V1)
			for k, v := range s1.seq {
				left[k] = append(left[k], v)
			}

			right := make(map[K][]V2)
			for k, v := range s2.seq {
				right[k] = append(right[k], v)
			}

			// Get all unique keys
			allKeys := make(map[K]struct{})
			for k := range left {
				allKeys[k] = struct{}{}
			}
			for k := range right {
				allKeys[k] = struct{}{}
			}

			// Yield grouped results
			for k := range allKeys {
				if !yield(CoGrouped[K, V1, V2]{
					Key:   k,
					Left:  left[k],
					Right: right[k],
				}) {
					return
				}
			}
		},
	}
}

// --- Stream-based Join (for Stream[T]) ---

// JoinBy performs an inner join on two streams using key extraction functions.
func JoinBy[T, U, K comparable](s1 Stream[T], s2 Stream[U], keyT func(T) K, keyU func(U) K) Stream[Pair[T, U]] {
	return Stream[Pair[T, U]]{
		seq: func(yield func(Pair[T, U]) bool) {
			// Build lookup from s2
			lookup := make(map[K][]U)
			for u := range s2.seq {
				k := keyU(u)
				lookup[k] = append(lookup[k], u)
			}

			// Join with s1
			for t := range s1.seq {
				k := keyT(t)
				if us, ok := lookup[k]; ok {
					for _, u := range us {
						if !yield(Pair[T, U]{First: t, Second: u}) {
							return
						}
					}
				}
			}
		},
	}
}

// LeftJoinBy performs a left join on two streams using key extraction functions.
func LeftJoinBy[T, U any, K comparable](s1 Stream[T], s2 Stream[U], keyT func(T) K, keyU func(U) K) Stream[Pair[T, Optional[U]]] {
	return Stream[Pair[T, Optional[U]]]{
		seq: func(yield func(Pair[T, Optional[U]]) bool) {
			lookup := make(map[K][]U)
			for u := range s2.seq {
				k := keyU(u)
				lookup[k] = append(lookup[k], u)
			}

			for t := range s1.seq {
				k := keyT(t)
				if us, ok := lookup[k]; ok {
					for _, u := range us {
						if !yield(Pair[T, Optional[U]]{First: t, Second: Some(u)}) {
							return
						}
					}
				} else {
					if !yield(Pair[T, Optional[U]]{First: t, Second: None[U]()}) {
						return
					}
				}
			}
		},
	}
}

// --- Anti-Join and Semi-Join ---

// SemiJoin returns elements from s1 that have matching keys in s2.
// Unlike inner join, it doesn't include the matching elements from s2.
func SemiJoin[K comparable, V1, V2 any](s1 Stream2[K, V1], s2 Stream2[K, V2]) Stream2[K, V1] {
	return Stream2[K, V1]{
		seq: func(yield func(K, V1) bool) {
			// Build set of keys from s2
			keys := make(map[K]struct{})
			for k := range s2.seq {
				keys[k] = struct{}{}
			}

			// Filter s1 to only matching keys
			for k, v := range s1.seq {
				if _, ok := keys[k]; ok {
					if !yield(k, v) {
						return
					}
				}
			}
		},
	}
}

// AntiJoin returns elements from s1 that don't have matching keys in s2.
func AntiJoin[K comparable, V1, V2 any](s1 Stream2[K, V1], s2 Stream2[K, V2]) Stream2[K, V1] {
	return Stream2[K, V1]{
		seq: func(yield func(K, V1) bool) {
			// Build set of keys from s2
			keys := make(map[K]struct{})
			for k := range s2.seq {
				keys[k] = struct{}{}
			}

			// Filter s1 to only non-matching keys
			for k, v := range s1.seq {
				if _, ok := keys[k]; !ok {
					if !yield(k, v) {
						return
					}
				}
			}
		},
	}
}

// SemiJoinBy returns elements from s1 that have matching keys in s2 (using key extractors).
func SemiJoinBy[T, U any, K comparable](s1 Stream[T], s2 Stream[U], keyT func(T) K, keyU func(U) K) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			keys := make(map[K]struct{})
			for u := range s2.seq {
				keys[keyU(u)] = struct{}{}
			}

			for t := range s1.seq {
				if _, ok := keys[keyT(t)]; ok {
					if !yield(t) {
						return
					}
				}
			}
		},
	}
}

// AntiJoinBy returns elements from s1 that don't have matching keys in s2 (using key extractors).
func AntiJoinBy[T, U any, K comparable](s1 Stream[T], s2 Stream[U], keyT func(T) K, keyU func(U) K) Stream[T] {
	return Stream[T]{
		seq: func(yield func(T) bool) {
			keys := make(map[K]struct{})
			for u := range s2.seq {
				keys[keyU(u)] = struct{}{}
			}

			for t := range s1.seq {
				if _, ok := keys[keyT(t)]; !ok {
					if !yield(t) {
						return
					}
				}
			}
		},
	}
}
