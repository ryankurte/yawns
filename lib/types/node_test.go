package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNodeType(t *testing.T) {

	t.Run("Links", func(t *testing.T) {
		t.Run("Filter", func(t *testing.T) {
			l := Links{Link{A: 0, B: 1, Fading: -100}, Link{A: 1, B: 0, Fading: -80}}
			d := l.Filter(func(l Link) bool {
				return l.A == 0 && l.B == 1
			})
			assert.Equal(t, Links{Link{A: 0, B: 1, Fading: -100}}, d)
		})
		t.Run("Map", func(t *testing.T) {
			l := Links{Link{A: 0, B: 1, Fading: -100}, Link{A: 1, B: 0, Fading: -80}}
			d := l.Map(func(v Link) Link {
				v.Fading = -90
				return v
			})
			assert.Equal(t, Links{Link{A: 0, B: 1, Fading: -90}, Link{A: 1, B: 0, Fading: -90}}, d)
		})
		t.Run("Deduplicate", func(t *testing.T) {
			l := Links{Link{A: 0, B: 1, Fading: -100}, Link{A: 1, B: 0, Fading: -80}}
			d := l.Deduplicate()
			assert.Equal(t, Links{Link{A: 0, B: 1, Fading: -90}}, d)
		})
		t.Run("Common", func(t *testing.T) {
			l1 := Links{Link{A: 0, B: 1, Fading: -100}, Link{A: 1, B: 0, Fading: -80}}
			l2 := Links{Link{A: 0, B: 1, Fading: -100}, Link{A: 2, B: 0, Fading: -80}}
			d := l1.Common(l2)
			assert.Equal(t, Links{Link{A: 0, B: 1, Fading: -100}}, d)
		})
		t.Run("Difference", func(t *testing.T) {
			l1 := Links{Link{A: 0, B: 1, Fading: -100}, Link{A: 1, B: 0, Fading: -80}}
			l2 := Links{Link{A: 0, B: 1, Fading: -100}, Link{A: 2, B: 0, Fading: -80}}
			d := l1.Difference(l2)
			assert.Equal(t, Links{Link{A: 1, B: 0, Fading: -80}, Link{A: 2, B: 0, Fading: -80}}, d)
		})

	})
}
