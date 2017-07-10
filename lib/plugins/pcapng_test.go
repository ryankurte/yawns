package plugins

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPCAPNG(t *testing.T) {

	t.Run("Encodes section header blocks", func(t *testing.T) {
		b := bytes.NewBuffer(nil)

		err := writeSectionHeader(b, nil)
		assert.Nil(t, err)

		expected := []byte{
			0x0A, 0x0D, 0x0D, 0x0A,
			0x1C, 0x00, 0x00, 0x00,
			0x4d, 0x3c, 0x2b, 0x1a,
			0x01, 0x00,
			0x00, 0x00,
			0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
			0x1C, 0x00, 0x00, 0x00,
		}
		assert.EqualValues(t, expected, b.Bytes())
	})

}
