package pi

import (
	"io"
	"strings"
	"testing"
)

func TestShortPromptIsTheArgument(t *testing.T) {
	args, stdin := invocation("fix the parser", false)
	if stdin != nil || args[0] != "-p" || args[1] != "fix the parser" {
		t.Fatalf("args=%v stdin=%v", args, stdin)
	}
}

// Over the Windows command-line limit the prompt travels on stdin, which pi merges in print mode.
func TestLongPromptTravelsOnStdin(t *testing.T) {
	long := strings.Repeat("x", maxArgMessage+1)
	args, stdin := invocation(long, false)
	if stdin == nil || args[1] != stdinCue {
		t.Fatalf("want the cue as -p and the prompt on stdin, got args=%v", args[:2])
	}
	got, _ := io.ReadAll(stdin)
	if string(got) != long {
		t.Fatal("stdin must carry the whole prompt")
	}
	if strings.Contains(strings.Join(args, " "), long) {
		t.Fatal("the long prompt must not be on the command line")
	}
}

// Through a launcher even a short prompt goes on stdin: the shell, wsl.exe and docker would
// re-quote an argument.
func TestLauncherPromptTravelsOnStdin(t *testing.T) {
	args, stdin := invocation("fix \"the\" parser\nnow", true)
	if stdin == nil || args[1] != stdinCue {
		t.Fatalf("want the cue as -p and the prompt on stdin, got args=%v", args)
	}
	got, _ := io.ReadAll(stdin)
	if string(got) != "fix \"the\" parser\nnow" {
		t.Fatalf("stdin=%q", got)
	}
}
