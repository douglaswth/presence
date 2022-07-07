package neighbors

type (
	State interface {
		Present() bool
		Changed() bool
		Set(present bool)
		Reset()
	}

	state struct {
		present, was, initial bool
	}
)

func NewState() State {
	return &state{initial: true}
}

func (s *state) Present() bool {
	return s.present
}

func (s *state) Changed() bool {
	return s.present != s.was
}

func (s *state) Set(present bool) {
	if s.initial {
		s.was = !present
		s.present = present
		s.initial = false
	} else {
		s.was = s.present
		s.present = present
	}
}

func (s *state) Reset() {
	s.initial = true
}
