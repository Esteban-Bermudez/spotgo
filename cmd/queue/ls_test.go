package queue

import "testing"

func TestTruncateForDisplay(t *testing.T) {
	cases := []struct {
		name  string
		input string
		w     int
		want  string
	}{
		{"shorter", "Hello", 10, "Hello"},
		{"exact", "HelloWorld", 10, "HelloWorld"},
		{"longer", "This is a very long title", 10, "This is..."},
		{"unicode", "日本語のタイトルが長いです", 10, "日本語のタイト..."},
		{"small width 3", "abcdef", 3, "abc"},
		{"width 1", "éåç", 1, "é"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := truncateForDisplay(c.input, c.w)
			if got != c.want {
				t.Fatalf("truncateForDisplay(%q, %d) = %q; want %q", c.input, c.w, got, c.want)
			}
		})
	}
}
