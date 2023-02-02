package log

import (
	"os"
	"testing"
)

func TestSetLevel(t *testing.T) {
	SetLevel(ErrorLevel)
	if infolog.Writer() == os.Stdout || errlog.Writer() != os.Stdout {
		t.Fatal("failed to set log level")
	}
	SetLevel(Disabled)
	if infolog.Writer() == os.Stdout || errlog.Writer() == os.Stdout {
		t.Fatal("failed to set log level")
	}
}
