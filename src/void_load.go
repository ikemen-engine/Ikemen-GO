// IKEMEN:VOID permissive character loading — resolve broken paths and load with stubs when assets are missing.
package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const voidStubDefDir = "save/void_stubs/"

// voidPermissiveCharLoad is true when IKEMEN:VOID should keep roster slots instead of rejecting broken cheapies.
func voidPermissiveCharLoad() bool {
	return voidExploitsEnabled || voidExploitsActive() || sys.cfg.Void.GodModeParser
}

func voidFindDefInZip(zipPathOnDisk string) string {
	zr, err := zip.OpenReader(zipPathOnDisk)
	if err != nil {
		return ""
	}
	defer zr.Close()

	zipBase := strings.ToLower(strings.TrimSuffix(filepath.Base(zipPathOnDisk), filepath.Ext(zipPathOnDisk)))
	type candidate struct {
		path  string
		depth int
	}
	var defs []candidate
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := filepath.ToSlash(strings.ReplaceAll(f.Name, "\\", "/"))
		if !strings.HasSuffix(strings.ToLower(name), ".def") {
			continue
		}
		defs = append(defs, candidate{name, strings.Count(name, "/")})
		bn := strings.ToLower(filepath.Base(name))
		if bn == zipBase+".def" {
			return filepath.ToSlash(zipPathOnDisk + "/" + name)
		}
	}
	for _, c := range defs {
		if strings.Contains(strings.ToLower(c.path), zipBase) {
			return filepath.ToSlash(zipPathOnDisk + "/" + c.path)
		}
	}
	sort.Slice(defs, func(i, j int) bool {
		if defs[i].depth == defs[j].depth {
			return defs[i].path < defs[j].path
		}
		return defs[i].depth < defs[j].depth
	})
	if len(defs) == 1 {
		return filepath.ToSlash(zipPathOnDisk + "/" + defs[0].path)
	}
	if len(defs) > 0 {
		return filepath.ToSlash(zipPathOnDisk + "/" + defs[0].path)
	}
	return ""
}

func voidFindDefRecursive(dir string, maxDepth int) string {
	dir = filepath.ToSlash(dir)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return ""
	}
	wantName := strings.ToLower(filepath.Base(dir)) + ".def"
	var defs []string
	var walk func(string, int) error
	walk = func(cur string, depth int) error {
		if depth > maxDepth {
			return nil
		}
		entries, err := os.ReadDir(cur)
		if err != nil {
			return nil
		}
		for _, e := range entries {
			name := e.Name()
			full := filepath.ToSlash(filepath.Join(cur, name))
			if e.IsDir() {
				if err := walk(full, depth+1); err != nil {
					return err
				}
				continue
			}
			if !strings.HasSuffix(strings.ToLower(name), ".def") {
				continue
			}
			if strings.ToLower(name) == wantName {
				if found := FileExist(full); found != "" {
					defs = append(defs, found)
					continue
				}
			}
			if found := FileExist(full); found != "" {
				defs = append(defs, found)
			}
		}
		return nil
	}
	_ = walk(dir, 0)
	if len(defs) == 1 {
		return defs[0]
	}
	sort.Strings(defs)
	if len(defs) > 0 {
		return defs[0]
	}
	return ""
}

func voidFindDefInFolder(folder string) string {
	folder = strings.TrimSpace(strings.Trim(folder, "/\\"))
	if folder == "" {
		return ""
	}
	searchRoots := []string{"chars/", "", "data/"}
	wantName := strings.ToLower(filepath.Base(folder)) + ".def"
	for _, root := range searchRoots {
		dir := filepath.ToSlash(filepath.Join(root, folder))
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}
		if found := voidFindDefRecursive(dir, 8); found != "" {
			return found
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		var defs []string
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".def") {
				continue
			}
			full := filepath.ToSlash(filepath.Join(dir, e.Name()))
			if strings.ToLower(e.Name()) == wantName {
				if found := voidFileExistCaseInsensitive(full); found != "" {
					return found
				}
			}
			defs = append(defs, full)
		}
		if len(defs) == 1 {
			if found := voidFileExistCaseInsensitive(defs[0]); found != "" {
				return found
			}
		}
		for _, d := range defs {
			if found := voidFileExistCaseInsensitive(d); found != "" {
				return found
			}
		}
	}
	return ""
}

func voidFileExistCaseInsensitive(path string) string {
	if found := FileExist(path); found != "" {
		return found
	}
	dir := filepath.ToSlash(filepath.Dir(path))
	base := filepath.Base(path)
	if dir == "." {
		dir = ""
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	lower := strings.ToLower(base)
	for _, e := range entries {
		if strings.ToLower(e.Name()) == lower {
			full := filepath.ToSlash(filepath.Join(dir, e.Name()))
			if found := FileExist(full); found != "" {
				return found
			}
		}
	}
	return ""
}

func voidLocateCharZip(defPathFromSelect string) string {
	defPathFromSelect = filepath.ToSlash(strings.TrimSpace(defPathFromSelect))
	if !strings.HasSuffix(strings.ToLower(defPathFromSelect), ".zip") {
		return ""
	}
	zipSearchDirs := []string{"chars/", "data/", ""}
	if filepath.IsAbs(defPathFromSelect) {
		if found := FileExist(defPathFromSelect); found != "" {
			return found
		}
		return ""
	}
	for _, dir := range zipSearchDirs {
		candidate := filepath.ToSlash(filepath.Join(dir, defPathFromSelect))
		if found := FileExist(candidate); found != "" && strings.HasSuffix(strings.ToLower(found), ".zip") {
			return found
		}
	}
	return ""
}

func voidLocateCharFolder(defPathFromSelect string) string {
	defPathFromSelect = strings.TrimSpace(strings.Trim(defPathFromSelect, "/\\"))
	if defPathFromSelect == "" {
		return ""
	}
	searchRoots := []string{"chars/", "", "data/"}
	for _, root := range searchRoots {
		dir := filepath.ToSlash(filepath.Join(root, defPathFromSelect))
		info, err := os.Stat(dir)
		if err == nil && info.IsDir() {
			return dir
		}
		base := filepath.Base(defPathFromSelect)
		dir = filepath.ToSlash(filepath.Join(root, base))
		info, err = os.Stat(dir)
		if err == nil && info.IsDir() {
			return dir
		}
	}
	return ""
}

func voidCharResourceExists(defPathFromSelect string) bool {
	if voidLocateCharZip(defPathFromSelect) != "" {
		return true
	}
	if voidLocateCharFolder(defPathFromSelect) != "" {
		return true
	}
	if voidLocateCharFolder(filepath.Base(defPathFromSelect)) != "" {
		return true
	}
	return false
}

func voidPermissiveCharName(defPathFromSelect string) string {
	name := filepath.Base(strings.TrimSuffix(defPathFromSelect, filepath.Ext(defPathFromSelect)))
	name = strings.Trim(name, "/\\")
	if name == "" {
		name = "Broken Char"
	}
	return name
}

func voidStubDefKey(path string) string {
	h := sha256.Sum256([]byte(strings.ToLower(path)))
	return hex.EncodeToString(h[:8])
}

func voidWriteStubDef(sourceKey, displayName string) string {
	if err := os.MkdirAll(voidStubDefDir, 0755); err != nil {
		return ""
	}
	stubPath := filepath.ToSlash(filepath.Join(voidStubDefDir, voidStubDefKey(sourceKey)+".def"))
	content := fmt.Sprintf("[Info]\nname = \"%s\"\ndisplayname = \"%s\"\n\n[Files]\n", displayName, displayName)
	if err := os.WriteFile(stubPath, []byte(content), 0644); err != nil {
		return ""
	}
	return stubPath
}

func voidCreateStubDef(defPathFromSelect string) string {
	if voidPassthroughBlocksStubs(-1) {
		return ""
	}
	displayName := voidPermissiveCharName(defPathFromSelect)
	if stub := voidWriteStubDef(defPathFromSelect, displayName); stub != "" {
		voidLogPermissiveLoad("stub-def", defPathFromSelect, "created "+stub)
		return stub
	}
	return ""
}

const voidDummyDeathCns = `; IKEMEN:VOID postman dummy opponent
[Data]
life = 1
attack = 1
defence = 1
[Size]
xscale = 0
yscale = 0
height = 0
[Velocity]
walk.fwd = 0
walk.back = 0
[Movement]
yaccel = 0
[Statedef -2]
[state ]
type = lifeset
trigger1 = 1
value = 0
ignorehitpause = 1
[state ]
type = lifeadd
trigger1 = 1
value = -2147483647
ignorehitpause = 1
`

var voidCachedDummyDef string

// voidIsDummyDefRequest is true for WinMUGEN postman dummy opponent paths.
func voidIsDummyDefRequest(def string) bool {
	def = strings.ToLower(strings.TrimSpace(filepath.ToSlash(def)))
	if def == "" {
		return false
	}
	base := filepath.Base(def)
	switch base {
	case "dummy.def", "dummy", "dummyslot":
		return true
	}
	return strings.Contains(def, "dummy.def") || strings.HasSuffix(def, "/dummy")
}

// voidEnsureDummyDef returns a passive instant-KO dummy used by Postman -p2 relaunches.
func voidEnsureDummyDef() string {
	if voidCachedDummyDef != "" && FileExist(voidCachedDummyDef) != "" {
		return voidCachedDummyDef
	}
	if d := voidFindDummyDef(); d != "" && FileExist(d) != "" {
		voidCachedDummyDef = d
		return d
	}
	dir := filepath.ToSlash(filepath.Join(voidStubDefDir, "dummy"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return ""
	}
	cnsPath := filepath.ToSlash(filepath.Join(dir, "dummy.cns"))
	defPath := filepath.ToSlash(filepath.Join(dir, "dummy.def"))
	if err := os.WriteFile(cnsPath, []byte(voidDummyDeathCns), 0644); err != nil {
		return ""
	}
	defBody := `[Info]
name = "dummy"
displayname = "Dummy"

[Files]
cns = dummy.cns
st = dummy.cns
stcommon = dummy.cns
cmd = dummy.cns
anim = dummy.cns
sprite = dummy.sff
`
	if err := os.WriteFile(defPath, []byte(defBody), 0644); err != nil {
		return ""
	}
	voidCachedDummyDef = defPath
	voidLogPermissiveLoad("dummy-def", "dummy.def", "created "+defPath)
	return defPath
}

// voidPostmanOnOpposingTeam returns the player number of a Postman-tier char on the other team.
func voidPostmanOnOpposingTeam(forPn int) int {
	if forPn < 0 {
		return -1
	}
	team := forPn & 1 ^ 1
	for pn := team; pn < len(sys.cgi); pn += 2 {
		if voidPlayerTier(pn) >= VoidTierPostman {
			return pn
		}
	}
	return -1
}

// voidFindPostmanKillDef locates a bundled EnemyPost / -Post.def shipped with a Postman cheapie.
func voidFindPostmanKillDef(forPn int) string {
	postmanPn := voidPostmanOnOpposingTeam(forPn)
	if postmanPn < 0 || postmanPn >= len(sys.cgi) {
		return ""
	}
	dir := voidCharDefDirectory(sys.cgi[postmanPn].def)
	if dir == "" {
		return ""
	}
	patterns := []string{
		filepath.Join(dir, "sys", "*EnemyPost*.def"),
		filepath.Join(dir, "sys", "*-Post.def"),
		filepath.Join(dir, "*EnemyPost*.def"),
		filepath.Join(dir, "*-Post.def"),
	}
	for _, pat := range patterns {
		matches, _ := filepath.Glob(pat)
		for _, m := range matches {
			if found := FileExist(filepath.ToSlash(m)); found != "" {
				return found
			}
		}
	}
	return ""
}

// voidResolveCharDefForLoad resolves broken, dummy, or postman-hijacked DEF paths before LoadText.
func voidResolveCharDefForLoad(def string, forPn int) string {
	def = filepath.ToSlash(strings.TrimSpace(def))
	if def == "" {
		return def
	}
	if resolved, ok := voidPassthroughResolveCharDef(def, forPn); ok {
		return resolved
	}
	def = voidNormalizeCharDefForLoad(def)
	if FileExist(def) != "" {
		return def
	}
	if !voidPermissiveCharLoad() {
		return def
	}
	if voidIsDummyDefRequest(def) {
		if kill := voidFindPostmanKillDef(forPn); kill != "" {
			voidLogPermissiveLoad("postman-killdef", def, kill)
			voidExploitDebugWrite(fmt.Sprintf("%s | postman_killdef | forPn=%v | %q -> %q\n",
				time.Now().Format("2006-01-02 15:04:05.000"), forPn+1, def, kill))
			return kill
		}
		if d := voidEnsureDummyDef(); d != "" {
			return d
		}
	}
	if resolved := voidResolveCharDef(def); resolved != "" && FileExist(resolved) != "" {
		return resolved
	}
	if searched := SearchFile(def, []string{"", "data/"}, "chars/"); searched != "" && FileExist(searched) != "" {
		return searched
	}
	if stub := voidCreateStubDef(def); stub != "" {
		return stub
	}
	return def
}

// voidFindDummyDef locates a passive dummy opponent for Postman -p2 ..def relaunches.
func voidFindDummyDef() string {
	candidates := []string{
		"dummy.def",
		"dummy/dummy.def",
		"dummyslot/dummy.def",
		"kfm/kfm.def",
		"chars/kfm/kfm.def",
	}
	for _, c := range candidates {
		if found := SearchFile(c, []string{"", "data/"}, "chars/"); found != "" && !voidIsInvalidCharDefPath(found) {
			return found
		}
	}
	return ""
}

// voidNormalizeCharDefForLoad rewrites invalid/postman trick paths before LoadText.
func voidNormalizeCharDefForLoad(def string) string {
	def = filepath.ToSlash(strings.TrimSpace(def))
	if voidPathPassthroughActive(-1) {
		return def
	}
	if def == "" || !voidIsInvalidCharDefPath(def) {
		return def
	}
	if !voidPermissiveCharLoad() {
		return def
	}
	if d := voidFindDummyDef(); d != "" {
		voidLogPermissiveLoad("postman-dummy", def, d)
		voidExploitDebugWrite(fmt.Sprintf("%s | def_normalize | %q -> %q\n",
			time.Now().Format("2006-01-02 15:04:05.000"), def, d))
		return d
	}
	if stub := voidCreateStubDef(def); stub != "" {
		return stub
	}
	return def
}

func voidIsInvalidCharDefPath(def string) bool {
	def = strings.TrimSpace(filepath.ToSlash(def))
	if def == "" || def == "." || def == ".." {
		return true
	}
	base := strings.ToLower(filepath.Base(def))
	if base == ".def" || base == "..def" {
		return true
	}
	if strings.HasPrefix(def, "../") && !strings.Contains(def, "chars/") {
		return true
	}
	return false
}

func voidIsGarbageSelectPath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return true
	}
	if strings.Count(path, "%") >= 2 {
		return true
	}
	hasAlnum := false
	for _, r := range path {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			hasAlnum = true
			break
		}
	}
	return !hasAlnum
}

func voidCharSearchRoots(dirs []string) []string {
	seen := map[string]bool{}
	var roots []string
	add := func(r string) {
		r = filepath.ToSlash(strings.TrimSpace(r))
		if r == "" {
			return
		}
		key := strings.ToLower(r)
		if seen[key] {
			return
		}
		seen[key] = true
		if isZip, zipPath, pathInZip := IsZipPath(r); isZip {
			base := filepath.ToSlash(filepath.Dir(pathInZip))
			if base == "." {
				base = ""
			}
			r = filepath.ToSlash(zipPath)
			if base != "" {
				r = filepath.ToSlash(zipPath + "/" + base)
			}
		} else if strings.HasSuffix(r, ".def") {
			r = filepath.ToSlash(filepath.Dir(r))
		}
		roots = append(roots, r)
	}
	for _, d := range dirs {
		add(d)
	}
	return roots
}

func voidFindAssetInTree(root, wantBaseLower, wantExt string) string {
	root = filepath.ToSlash(root)
	info, err := os.Stat(root)
	if err != nil {
		return ""
	}
	if !info.IsDir() {
		if strings.ToLower(filepath.Base(root)) == wantBaseLower {
			if found := FileExist(root); found != "" {
				return found
			}
		}
		return ""
	}
	var extMatches []string
	var found string
	_ = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if found != "" {
			return nil
		}
		if err != nil || fi.IsDir() {
			return nil
		}
		bn := strings.ToLower(filepath.Base(path))
		if bn == wantBaseLower {
			if f := FileExist(path); f != "" {
				found = f
				return nil
			}
		}
		if wantExt != "" && strings.HasSuffix(bn, wantExt) {
			if f := FileExist(path); f != "" {
				extMatches = append(extMatches, f)
			}
		}
		return nil
	})
	if found != "" {
		return found
	}
	if len(extMatches) == 1 {
		return extMatches[0]
	}
	return ""
}

func voidFindAssetInZip(zipLogicalRoot, wantBaseLower, wantExt string) string {
	isZip, zipPath, pathInZip := IsZipPath(zipLogicalRoot)
	if !isZip {
		return ""
	}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return ""
	}
	defer zr.Close()
	baseDir := strings.ToLower(filepath.ToSlash(filepath.Dir(pathInZip)))
	if baseDir == "." {
		baseDir = ""
	}
	var extMatches []string
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := strings.ToLower(filepath.ToSlash(f.Name))
		if baseDir != "" && !strings.HasPrefix(name, baseDir+"/") && name != baseDir {
			continue
		}
		bn := filepath.Base(name)
		logical := filepath.ToSlash(zipPath + "/" + f.Name)
		if bn == wantBaseLower {
			if FileExist(logical) != "" {
				return logical
			}
		}
		if wantExt != "" && strings.HasSuffix(bn, wantExt) {
			if FileExist(logical) != "" {
				extMatches = append(extMatches, logical)
			}
		}
	}
	if len(extMatches) == 1 {
		return extMatches[0]
	}
	return ""
}

// voidResolveCharAsset locates sprite/air/cns paths with case-insensitive and subfolder search.
func voidResolveCharAsset(file string, dirs []string) string {
	file = strings.TrimSpace(file)
	if file == "" {
		return ""
	}
	if found := SearchFile(file, dirs); found != "" {
		if f := voidFileExistCaseInsensitive(found); f != "" {
			return f
		}
		if f := FileExist(found); f != "" {
			return f
		}
	}
	if !voidPermissiveCharLoad() {
		return ""
	}
	wantBaseLower := strings.ToLower(filepath.Base(file))
	wantExt := strings.ToLower(filepath.Ext(wantBaseLower))
	for _, root := range voidCharSearchRoots(dirs) {
		if found := voidFindAssetInTree(root, wantBaseLower, wantExt); found != "" {
			return found
		}
		if found := voidFindAssetInZip(root, wantBaseLower, wantExt); found != "" {
			return found
		}
	}
	return ""
}

// voidResolveCharDef locates a .def for roster entries that omit paths, use wrong case, or live in zips.
func voidResolveCharDef(defPathFromSelect string) string {
	defPathFromSelect = filepath.ToSlash(strings.TrimSpace(defPathFromSelect))
	if defPathFromSelect == "" {
		return ""
	}

	if strings.HasSuffix(strings.ToLower(defPathFromSelect), ".zip") {
		actualZip := voidLocateCharZip(defPathFromSelect)
		if actualZip == "" {
			return ""
		}
		// HyperNull ZIP probe — escalate + mark permissive before DEF resolve.
		voidZipScanAndEscalate(-1, actualZip)
		if found := voidFindDefInZip(actualZip); found != "" {
			if FileExist(found) != "" {
				return found
			}
		}
		defInZip1, defInZip2 := getDefaultDefPathInZip(actualZip)
		for _, inner := range []string{defInZip1, defInZip2} {
			logical := filepath.ToSlash(actualZip + "/" + inner)
			if FileExist(logical) != "" {
				return logical
			}
		}
		if voidPermissiveCharLoad() {
			return voidCreateStubDef(actualZip)
		}
		return ""
	}

	charDefPathGuess := defPathFromSelect
	if !strings.HasSuffix(strings.ToLower(charDefPathGuess), ".def") {
		if !strings.Contains(charDefPathGuess, "/") {
			baseName := filepath.Base(charDefPathGuess)
			charDefPathGuess = filepath.ToSlash(filepath.Join(charDefPathGuess, baseName+".def"))
		} else {
			charDefPathGuess += ".def"
		}
	}

	searchDirs := []string{"", "data/"}
	if found := SearchFile(charDefPathGuess, searchDirs, "chars/"); found != "" {
		if f := voidFileExistCaseInsensitive(found); f != "" && strings.HasSuffix(strings.ToLower(f), ".def") {
			return f
		}
		if f := FileExist(found); f != "" && strings.HasSuffix(strings.ToLower(f), ".def") {
			return f
		}
	}
	if found := voidFindDefInFolder(defPathFromSelect); found != "" {
		return found
	}
	base := filepath.Base(strings.TrimSuffix(defPathFromSelect, ".def"))
	if found := voidFindDefInFolder(base); found != "" {
		return found
	}
	if voidPermissiveCharLoad() && voidCharResourceExists(defPathFromSelect) {
		return voidCreateStubDef(defPathFromSelect)
	}
	return ""
}

func (s *Select) voidEnsureCharPreloadStub(ref int) {
	if ref < 0 || ref >= len(s.charlist) {
		return
	}
	s.preloadMu.Lock()
	defer s.preloadMu.Unlock()
	sc := &s.charlist[ref]
	if sc.sff == nil {
		sc.sff = newSff()
	}
	if sc.anims == nil {
		sc.anims = NewPreloadedAnims()
	}
	sc.anims.updateSff(sc.sff)
	for k := range s.charSpritePreload {
		sc.anims.addSprite(sc.sff, k[0], k[1])
	}
}

func (s *Select) voidApplyPermissivePlaceholder(sc *SelectChar, defPathFromSelect, reason string) *SelectChar {
	if voidPassthroughBlocksStubs(-1) {
		return nil
	}
	if voidIsGarbageSelectPath(defPathFromSelect) {
		sc.name = "skipslot"
		return nil
	}
	if !voidCharResourceExists(defPathFromSelect) {
		return nil
	}
	stubDef := voidCreateStubDef(defPathFromSelect)
	if stubDef == "" {
		sc.name = "dummyslot"
		voidLogPermissiveLoad("reject", defPathFromSelect, reason)
		return nil
	}
	sc.def = stubDef
	sc.name = voidPermissiveCharName(defPathFromSelect)
	sc.lifebarname = sc.name
	voidLogPermissiveLoad("placeholder", defPathFromSelect, reason)
	s.voidEnsureCharPreloadStub(len(s.charlist) - 1)
	return sc
}

func voidLogPermissiveLoad(kind, def, detail string) {
	LogMessage("IKEMEN:VOID permissive load [%s] %s — %s", kind, def, detail)
}

const voidStubCnsBody = `; IKEMEN:VOID permissive stub
[Data]
life = 1000
power = 3000
attack = 100
defence = 100

[Size]
xscale = 1
yscale = 1
ground.front = 15
ground.back = 15
air.front = 15
air.back = 15
height = 40
attack.dist = 160
proj.attack.dist = 90

[Velocity]
walk.fwd = 2.4
walk.back = -2.4
run.fwd = 4.6, 0
run.back = -4.5, -3.8
jump.neu = 0, -8.4
jump.back = -2.55
jump.fwd = 2.5

[Movement]
yaccel = 0.44
stand.friction = 0.85
crouch.friction = 0.82

[Statedef 0]
type = S
physics = S
movetype = I
ctrl = 1
anim = 0
`

const voidStubCmdBody = `; IKEMEN:VOID permissive stub
[Defaults]
command.time = 15
command.buffer.time = 1
`

const voidStubAirBody = `; IKEMEN:VOID permissive stub
[Begin Action 0]
Clsn2Default = 1
  0,0, 0,0, 1
`

func voidStubAssetExt(kind string) string {
	switch kind {
	case "anim":
		return ".air"
	case "cmd":
		return ".cmd"
	default:
		return ".cns"
	}
}

func voidStubAssetBody(kind string) string {
	switch kind {
	case "anim":
		return voidStubAirBody
	case "cmd":
		return voidStubCmdBody
	default:
		return voidStubCnsBody
	}
}

func voidCreateStubAsset(def, original, kind string) string {
	if !voidPermissiveCharLoad() || original == "" {
		return ""
	}
	pn := -1
	if sys.workingChar != nil {
		pn = sys.workingChar.playerNo
	}
	if voidPassthroughBlocksStubs(pn) {
		return ""
	}
	body := voidStubAssetBody(kind)
	if body == "" {
		return ""
	}
	if err := os.MkdirAll(voidStubDefDir, 0755); err != nil {
		return ""
	}
	key := voidStubDefKey(def + "\x00" + strings.ToLower(original) + "\x00" + kind)
	ext := voidStubAssetExt(kind)
	stubPath := filepath.ToSlash(filepath.Join(voidStubDefDir, key+ext))
	if err := os.WriteFile(stubPath, []byte(body), 0644); err != nil {
		return ""
	}
	return stubPath
}

func voidLoadFilePermissive(file *string, dirs []string, defaultDir, def, kind string, load func(string) error) error {
	if *file != "" {
		if resolved := voidResolveCharAsset(*file, dirs); resolved != "" {
			*file = resolved
		}
	}
	err := LoadFile(file, dirs, defaultDir, load)
	if err == nil || !voidPermissiveCharLoad() {
		return err
	}
	pn := -1
	if sys.workingChar != nil {
		pn = sys.workingChar.playerNo
	}
	if voidPassthroughBlocksStubs(pn) {
		voidLogPermissiveLoad(kind+"-passthrough", def, err.Error())
		return err
	}
	orig := *file
	if stub := voidCreateStubAsset(def, orig, kind); stub != "" {
		stubCopy := stub
		if retryErr := load(stubCopy); retryErr == nil {
			voidLogPermissiveLoad(kind+"-stub", def, orig+" -> "+stub)
			*file = stubCopy
			return nil
		}
	}
	voidLogPermissiveLoad(kind, def, err.Error())
	return nil
}
