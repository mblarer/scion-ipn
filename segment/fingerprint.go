package segment

import (
	"github.com/mblarer/scion-ipn/path"
	"github.com/scionproto/scion/go/lib/snet"
)

// Fingerprint creates a string that uniquely identifies a segment with high
// probability solely based on the sequence of its path interfaces.
func Fingerprint(segment Segment) string {
	path := path.InterfacePath{segment.PathInterfaces()}
	fingerprint := snet.Fingerprint(path)
	return string(fingerprint)
}
