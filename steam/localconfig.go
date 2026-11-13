package steam

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Rethunk-Tech/go-gamelocator/internal/readcap"
	"github.com/Rethunk-Tech/go-safefs"
)

// vdfToken is one token of a text VDF file with its byte span, so an edit can change one value and leave every
// other byte of Steam's file as it was.
type vdfToken struct {
	kind       byte // 's' quoted string, '{', '}'
	text       string
	start, end int
}

func scanVDF(b []byte) ([]vdfToken, error) {
	var out []vdfToken
	for i := 0; i < len(b); {
		c := b[i]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			i++
		case c == '/' && i+1 < len(b) && b[i+1] == '/':
			for i < len(b) && b[i] != '\n' {
				i++
			}
		case c == '{' || c == '}':
			out = append(out, vdfToken{kind: c, start: i, end: i + 1})
			i++
		case c == '"':
			start := i
			var sb strings.Builder
			i++
			for i < len(b) && b[i] != '"' {
				if b[i] == '\\' && i+1 < len(b) {
					i++
					switch b[i] {
					case 'n':
						sb.WriteByte('\n')
					case 't':
						sb.WriteByte('\t')
					default:
						sb.WriteByte(b[i])
					}
				} else {
					sb.WriteByte(b[i])
				}
				i++
			}
			if i >= len(b) {
				return nil, errors.New("unterminated string in VDF")
			}
			i++
			out = append(out, vdfToken{kind: 's', text: sb.String(), start: start, end: i})
		default:
			return nil, fmt.Errorf("unexpected %q in VDF at byte %d", c, i)
		}
	}
	return out, nil
}

func quoteVDF(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

// setLaunchOptions sets apps/<appID>/LaunchOptions under UserLocalConfigStore/Software/Valve/Steam to merge(old),
// creating the app's block or the key when missing; it returns the new file and the value it wrote.
func setLaunchOptions(b []byte, appID string, merge func(old string) string) ([]byte, string, error) {
	toks, err := scanVDF(b)
	if err != nil {
		return nil, "", err
	}
	loc, err := locateLaunchOptions(toks, appID)
	if err != nil {
		return nil, "", err
	}
	appsClose, appOpen, appClose, optVal := loc.appsClose, loc.appOpen, loc.appClose, loc.optVal
	switch {
	case optVal != nil:
		value := merge(optVal.text)
		return splice(b, optVal.start, optVal.end, quoteVDF(value)), value, nil
	case appOpen >= 0 && appClose >= 0:
		value := merge("")
		return splice(b, appClose, appClose, "\t\"LaunchOptions\"\t\t"+quoteVDF(value)+"\n\t\t\t\t\t"), value, nil
	case appsClose >= 0:
		value := merge("")
		block := quoteVDF(appID) + "\n\t\t\t\t\t{\n\t\t\t\t\t\t\"LaunchOptions\"\t\t" + quoteVDF(value) + "\n\t\t\t\t\t}\n\t\t\t\t"
		return splice(b, appsClose, appsClose, block), value, nil
	}
	return nil, "", ErrNoAppsSection
}

// launchLoc is where setLaunchOptions edits: the byte ranges of the apps block, the app's block and its LaunchOptions
// value, with -1 or nil for any the file lacks.
type launchLoc struct {
	appsClose, appOpen, appClose int
	optVal                       *vdfToken
}

// launchOptionsWant is the key path of the apps block holding one block per app id.
var launchOptionsWant = []string{"userlocalconfigstore", "software", "valve", "steam", "apps"}

// vdfPath tracks the key path while locateLaunchOptions walks the tokens.
type vdfPath struct {
	path    []string
	pending string
	appID   string
	loc     launchLoc
}

func (p *vdfPath) inApp() bool {
	want := launchOptionsWant
	return len(p.path) == len(want)+1 && pathIs(p.path, want) && p.path[len(want)] == p.appID
}

// locateLaunchOptions walks toks tracking the key path and finds the edit sites for appID.
func locateLaunchOptions(toks []vdfToken, appID string) (launchLoc, error) {
	p := &vdfPath{appID: appID, loc: launchLoc{appsClose: -1, appOpen: -1, appClose: -1}}
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		switch t.kind {
		case 's':
			if p.pending == "" {
				i += p.key(toks, i)
			}
		case '{':
			p.open(t)
		case '}':
			if len(p.path) == 0 {
				return launchLoc{}, errors.New("unbalanced braces in VDF")
			}
			p.close(t)
		}
	}
	return p.loc, nil
}

// key handles the string token at toks[i], returning how many following tokens it consumed: a key with a string value
// is a pair, and the LaunchOptions pair of the app is the value to replace.
func (p *vdfPath) key(toks []vdfToken, i int) int {
	t := toks[i]
	p.pending = t.text
	if i+1 < len(toks) && toks[i+1].kind == 's' {
		if p.inApp() && strings.EqualFold(t.text, "LaunchOptions") {
			p.loc.optVal = &toks[i+1]
		}
		p.pending = ""
		return 1
	}
	return 0
}

func (p *vdfPath) open(t vdfToken) {
	want := launchOptionsWant
	p.path = append(p.path, strings.ToLower(p.pending))
	if pathIs(p.path, want) && len(p.path) == len(want)+1 && p.pending == p.appID {
		p.loc.appOpen = t.end
	}
	p.pending = ""
}

func (p *vdfPath) close(t vdfToken) {
	want := launchOptionsWant
	if p.inApp() {
		p.loc.appClose = t.start
	}
	if len(p.path) == len(want) && pathIs(p.path, want) {
		p.loc.appsClose = t.start
	}
	p.path = p.path[:len(p.path)-1]
}

func pathIs(path, want []string) bool {
	if len(path) < len(want) {
		return false
	}
	for i, w := range want {
		if path[i] != w {
			return false
		}
	}
	return true
}

func splice(b []byte, start, end int, s string) []byte {
	out := make([]byte, 0, len(b)+len(s))
	out = append(out, b[:start]...)
	out = append(out, s...)
	return append(out, b[end:]...)
}

// SetLaunchOptions changes appID's launch options for the MostRecent account to merge(current). Steam rewrites
// localconfig.vdf on exit, so the caller makes sure Steam is closed. It returns the value written.
func (s Steam) SetLaunchOptions(appID string, merge func(current string) string, o EditOptions) (string, error) {
	if err := o.validate(); err != nil {
		return "", err
	}
	dir, err := s.userConfigDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "localconfig.vdf")
	b, err := readcap.File(path, readcap.Config)
	if err != nil {
		return "", err
	}
	next, value, err := setLaunchOptions(b, appID, merge)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if err := backupBeforeEdit(path, o.BackupSuffix, info.Mode().Perm()); err != nil {
		return "", err
	}
	if err := safefs.WriteFileAtomic(path, next, info.Mode().Perm()); err != nil {
		return "", err
	}
	// The atomic write lets the umask narrow the new file; Steam's file keeps the mode it had.
	return value, safefs.Chmod(path, info.Mode().Perm())
}
