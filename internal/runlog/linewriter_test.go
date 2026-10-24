package runlog

import (
	"reflect"
	"strings"
	"testing"
)

func collect() (*LineWriter, *[]string) {
	var lines []string
	return NewLineWriter(func(l string) { lines = append(lines, l) }), &lines
}

func TestLineWriter(t *testing.T) {
	t.Run("splits writes into lines", func(t *testing.T) {
		w, lines := collect()
		_, _ = w.Write([]byte("one\ntw"))
		_, _ = w.Write([]byte("o\nthree\n"))
		w.Flush()
		if want := []string{"one", "two", "three"}; !reflect.DeepEqual(*lines, want) {
			t.Errorf("lines = %q, want %q", *lines, want)
		}
	})

	t.Run("flush emits a final line without newline", func(t *testing.T) {
		w, lines := collect()
		_, _ = w.Write([]byte("done"))
		if len(*lines) != 0 {
			t.Fatalf("emitted %q before Flush", *lines)
		}
		w.Flush()
		w.Flush()
		if want := []string{"done"}; !reflect.DeepEqual(*lines, want) {
			t.Errorf("lines = %q, want %q", *lines, want)
		}
	})

	t.Run("keeps empty lines and strips CR", func(t *testing.T) {
		w, lines := collect()
		_, _ = w.Write([]byte("a\r\n\nb\n"))
		if want := []string{"a", "", "b"}; !reflect.DeepEqual(*lines, want) {
			t.Errorf("lines = %q, want %q", *lines, want)
		}
	})

	t.Run("splits overlong lines", func(t *testing.T) {
		w, lines := collect()
		_, _ = w.Write([]byte(strings.Repeat("x", MaxLineBytes+10) + "\n"))
		if len(*lines) != 2 || len((*lines)[0]) != MaxLineBytes || len((*lines)[1]) != 10 {
			t.Errorf("got %d lines of lengths %v", len(*lines), lens(*lines))
		}
	})

	t.Run("splits overlong output with no newline before Flush", func(t *testing.T) {
		w, lines := collect()
		_, _ = w.Write([]byte(strings.Repeat("y", 2*MaxLineBytes+1)))
		if len(*lines) != 2 {
			t.Fatalf("got %d lines before Flush, want 2", len(*lines))
		}
		w.Flush()
		if len(*lines) != 3 || len((*lines)[2]) != 1 {
			t.Errorf("got lengths %v after Flush", lens(*lines))
		}
	})
}

func lens(lines []string) []int {
	n := make([]int, len(lines))
	for i, l := range lines {
		n[i] = len(l)
	}
	return n
}
