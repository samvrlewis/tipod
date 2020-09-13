package encoding

import (
	"bytes"
	"testing"
)

func TestToDomainPrefix(t *testing.T) {
	input := []byte("tonsnandtonsofbytesallinarowwhenwilltheystopnobodyknowsoktheymayaswellstopnow")

	encoded, err := ToDomainPrefix(input)

	if err != nil {
		t.Error("Unexpected error: ", err)
	}

	decoded, err := FromDomainPrefix(encoded)

	if err != nil {
		t.Error("Unexpected error: ", err)
	}

	if !bytes.Equal(input, decoded) {
		t.Error("Expected decode to input")
	}

}
