package manifest

// Mixins is set of mixins
type Mixins map[string]Mixin

type Mixin struct {
	JobsContainer
}

func newMixin(container JobsContainer) Mixin {
	return Mixin{
		JobsContainer: container,
	}
}
