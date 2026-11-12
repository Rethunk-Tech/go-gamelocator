package readcap

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Rethunk-Tech/go-safefs"
)

func TestFile(t *testing.T) {
	root := t.TempDir()
	put := func(name string, n int) string {
		t.Helper()
		p := filepath.Join(root, name)
		if err := os.WriteFile(p, bytes.Repeat([]byte{'x'}, n), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	dir := filepath.Join(root, "dir")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}

	const limit int64 = 4
	for _, tc := range []struct {
		name    string
		path    string
		want    int
		wantErr error
		anyErr  bool
	}{
		{name: "under", path: put("under", 3), want: 3},
		{name: "exact", path: put("exact", 4), want: 4},
		{name: "over", path: put("over", 5), wantErr: safefs.ErrTooLarge},
		{name: "missing", path: filepath.Join(root, "gone"), wantErr: os.ErrNotExist},
		{name: "not regular", path: dir, anyErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := File(tc.path, limit)
			switch {
			case tc.anyErr:
				if err == nil {
					t.Fatal("want error")
				}
			case !errors.Is(err, tc.wantErr):
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			if len(got) != tc.want {
				t.Fatalf("len = %d, want %d", len(got), tc.want)
			}
		})
	}
}
