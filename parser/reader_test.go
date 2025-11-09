package parser

import "testing"

func TestStreamReadPeek(t *testing.T) {
	r := newReader([]byte("hello world"))

	if ch, err := r.peek(); err != nil || ch != 'h' {
		t.Fatalf("peek: expected 'h', got %q (err=%v)", ch, err)
	}

	if ch1, ch2, err := r.peek2(); err != nil || ch1 != 'h' || ch2 != 'e' {
		t.Fatalf("peek2: expected 'h','e', got %q,%q (err=%v)", ch1, ch2, err)
	}

	if buf, err := r.peekN(3); err != nil || string(buf) != "hel" {
		t.Fatalf("peekN(3): expected \"hel\", got %q (err=%v)", buf, err)
	}

	if err := r.discard(3); err != nil {
		t.Fatalf("discard: %v", err)
	}

	if ch, err := r.peek(); err != nil || ch != 'l' {
		t.Fatalf("peek after discard: expected 'l', got %q (err=%v)", ch, err)
	}

	if ch1, ch2, err := r.peek2(); err != nil || ch1 != 'l' || ch2 != 'o' {
		t.Fatalf("peek2 after discard: expected 'l','o', got %q,%q (err=%v)", ch1, ch2, err)
	}

	if buf, err := r.peekN(3); err != nil || string(buf) != "lo " {
		t.Fatalf("peekN(3) after discard: expected \"lo \", got %q (err=%v)", buf, err)
	}
}
