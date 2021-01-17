package alpaca

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscoveryMessageString(t *testing.T) {
	adm := NewAlpacaDiscoveryMessage(ALPACA_DISCOVERY_VERSION)
	assert.Equal(t, "alpacadiscovery1", adm.String())

	adm = NewAlpacaDiscoveryMessage(5)
	assert.Equal(t, "alpacadiscovery5", adm.String())
}
