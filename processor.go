package ta

// Processor consumes values one at a time and emits a computed value once ready.
//
// Indicators return false from Process while warming up (e.g. until a moving
// average window is full) and true thereafter. Callers should ignore results
// paired with false.
type Processor[I any, O any] interface {
	Process(v I) (O, bool)
}
