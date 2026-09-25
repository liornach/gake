package logic

import (
	"bytes"
	"strings"
	"testing"
)

// -h prints the usage (listing both flags), succeeds, and touches nothing.
func TestRunGakeHelp(t *testing.T) {
	for _, helpFlag := range []string{"-h", "--help"} {
		t.Run(helpFlag, func(t *testing.T) {
			root := createStubCopy(t, "new-c")
			before := snapshotDir(t, root)
			var out bytes.Buffer

			if err := runGake(root, []string{helpFlag}, &out); err != nil {
				t.Fatalf("unexpected err %v", err)
			}

			for _, flagName := range []string{"-projname", "-lang"} {
				if !strings.Contains(out.String(), flagName) {
					t.Fatalf("usage does not mention %s:\n%s", flagName, out.String())
				}
			}
			assertDirUnchanged(t, root, before)
		})
	}
}
