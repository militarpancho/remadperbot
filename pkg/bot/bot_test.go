package bot

import "testing"

func TestArticleURLUsesCurrentRemadDetailRoute(t *testing.T) {
	got := articleURL("abc123")
	want := "https://remad.madrid.es/REMAD_FTP/#/detalleAntique/abc123"

	if got != want {
		t.Fatalf("articleURL = %q, want %q", got, want)
	}
}
