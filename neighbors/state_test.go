package neighbors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	s := NewState()
	assert.Equal(t, &state{present: false, was: false, initial: true}, s)
}

func TestState_Present(t *testing.T) {
	cases := []struct {
		name string
		s    State
		exp  bool
	}{
		{
			name: "true",
			s:    &state{present: true},
			exp:  true,
		},
		{
			name: "false",
			s:    &state{present: false},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.exp, tc.s.Present())
		})
	}
}

func TestState_Changed(t *testing.T) {
	cases := []struct {
		name string
		s    State
		exp  bool
	}{
		{
			name: "true to true",
			s:    &state{present: true, was: true},
			exp:  false,
		},
		{
			name: "true to false",
			s:    &state{present: false, was: true},
			exp:  true,
		},
		{
			name: "false to true",
			s:    &state{present: true, was: false},
			exp:  true,
		},
		{
			name: "false to false",
			s:    &state{present: false, was: false},
			exp:  false,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.exp, tc.s.Changed())
		})
	}
}

func TestState_Set(t *testing.T) {
	cases := []struct {
		name   string
		s, exp State
		p      bool
	}{
		{
			name: "initial to true",
			s:    &state{initial: true},
			p:    true,
			exp:  &state{present: true, was: false, initial: false},
		},
		{
			name: "initial to false",
			s:    &state{initial: true},
			p:    false,
			exp:  &state{present: false, was: true, initial: false},
		},
		{
			name: "true to true",
			s:    &state{present: true},
			p:    true,
			exp:  &state{present: true, was: true},
		},
		{
			name: "true to false",
			s:    &state{present: true},
			p:    false,
			exp:  &state{present: false, was: true},
		},
		{
			name: "false to true",
			s:    &state{present: false, was: true},
			p:    true,
			exp:  &state{present: true, was: false},
		},
		{
			name: "false to false",
			s:    &state{present: false, was: true},
			p:    false,
			exp:  &state{present: false, was: false},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.s.Set(tc.p)
			assert.Equal(t, tc.exp, tc.s)
		})
	}
}

func TestState_Reset(t *testing.T) {
	s := &state{initial: false}
	s.Reset()
	assert.Equal(t, &state{initial: true}, s)
}
