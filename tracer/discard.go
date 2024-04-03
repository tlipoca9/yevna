package tracer

// Discard is a [Tracer] on which all Trace calls succeed
// without doing anything.
var Discard Tracer = discard{}

type discard struct{}

func (discard) Trace(string, ...string) {}
