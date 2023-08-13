package mmock

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
)

// MockGenerateFile generates a Go file with a mock for interface type T
//
// The pkg arg is the package name (or package path) for the code - if this is an empty
// string then the package of the interface type T is used
//
// Caveat emptor! The generated code is designed to save you work but is not 100% guaranteed to produce
// compilable code (it can get confused with convoluted or conflicting package names) and may sometimes
// require manual intervention.
func MockGenerateFile[T any](pkg string, f *os.File) error {
	w := newWriter(bufio.NewWriter(f))
	mockGenerate[T](pkg, w)
	return w.flush()
}

// MockGenerate generates Go code with a mock for interface type T
//
// The pkg arg is the package name (or package path) for the code - if this is an empty
// string then the package of the interface type T is used
//
// Caveat emptor! The generated code is designed to save you work but is not 100% guaranteed to produce
// compilable code (it can get confused with convoluted or conflicting package names) and may sometimes
// require manual intervention.
func MockGenerate[T any](pkg string) ([]byte, error) {
	w := newWriter(nil)
	mockGenerate[T](pkg, w)
	return w.bytes()
}

func mockGenerate[T any](pkg string, w *writer) {
	def := newMockDef[T](pkg)
	def.write(w)
}

type packages map[string]struct{}

func newPackages(initial ...string) packages {
	r := packages{}
	for _, s := range initial {
		r[s] = struct{}{}
	}
	return r
}

func (ps packages) addPackages(add packages) {
	for pkg := range add {
		ps[pkg] = struct{}{}
	}
}

func (ps packages) add(pkg string) {
	ps[pkg] = struct{}{}
}

func (ps packages) writeImports(w *writer, defPkg string) {
	imps := make([]string, 0, len(ps))
	for imp := range ps {
		if imp != "" {
			if pkg, core := packagePathToPackage(imp); core || pkg != defPkg {
				imps = append(imps, "\t\""+imp+"\"")
			}
		}
	}
	if len(imps) > 0 {
		w.writeLines("import (")
		sort.Strings(imps)
		w.writeLines(imps...)
		w.writeLines(")", "")
	}
}

type mockDef struct {
	fns  []mockFunc
	pkgs packages
	intf string
	name string
	pkg  string
}

const (
	pkgMmock       = "github.com/go-andiamo/mmock"
	mockNamePrefix = "Mock"
)

func newMockDef[T any](pkg string) mockDef {
	r := &mockDef{
		fns:  []mockFunc{},
		pkgs: newPackages(pkgMmock),
	}
	tt := reflect.TypeOf((*T)(nil)).Elem()
	if pkg == "" {
		r.pkg, _ = packagePathToPackage(tt.PkgPath())
	} else {
		r.pkg, _ = packagePathToPackage(pkg)
	}
	r.intf = tt.Name()
	r.name = mockNamePrefix + tt.Name()
	for i := 0; i < tt.NumMethod(); i++ {
		fn := newMockFunc(tt.Method(i), r.pkg)
		r.fns = append(r.fns, fn)
		r.pkgs.addPackages(fn.pkgs)
	}
	return *r
}

func (m mockDef) write(w *writer) {
	w.writeLines("package "+m.pkg, "")
	m.pkgs.writeImports(w, m.pkg)
	w.writeLines(
		"type "+m.name+" struct {",
		"\tmmock.MockMethods",
		"}", "",
		"func New"+m.name+"() *"+m.name+" {",
		"\treturn mmock.NewMockOf["+m.name+","+m.intf+"]()",
		"}", "",
		"// make sure mock implements interface...",
		"var _ "+m.intf+" = &"+m.name+"{}")
	for _, fn := range m.fns {
		fn.write(w, m.name)
	}
}

type mockFunc struct {
	pkg       string
	name      string
	pkgs      packages
	ins       []mockArg
	outs      []mockArg
	isVaradic bool
}

func newMockFunc(m reflect.Method, pkg string) mockFunc {
	mt := m.Type
	r := &mockFunc{
		pkg:       pkg,
		name:      m.Name,
		pkgs:      newPackages(),
		ins:       []mockArg{},
		outs:      []mockArg{},
		isVaradic: mt.IsVariadic(),
	}
	inLen := mt.NumIn()
	for i := 0; i < inLen; i++ {
		a := newMockArg(r.pkg, mt.In(i), r.isVaradic && i == inLen-1)
		r.pkgs.add(a.pkg)
		r.ins = append(r.ins, a)
	}
	for i := 0; i < mt.NumOut(); i++ {
		a := newMockArg(r.pkg, mt.Out(i), false)
		r.pkgs.add(a.pkg)
		r.outs = append(r.outs, a)
	}
	return *r
}

func (f mockFunc) write(w *writer, receiver string) {
	w.writeLines("")
	w.write("func (m *" + receiver + ") " + f.name)
	callArgs := f.writeCallArgs(w)
	f.writeOutArgs(w)
	f.writeBody(callArgs, w)
}

func (f mockFunc) writeCallArgs(w *writer) []string {
	callArgs := make([]string, 0, len(f.ins))
	w.write("(")
	for i, a := range f.ins {
		if i > 0 {
			w.write(", ")
		}
		argName := fmt.Sprintf("arg%d", i+1)
		callArgs = append(callArgs, argName)
		w.write(fmt.Sprintf("%s %s", argName, a.fullName()))
	}
	w.write(")")
	return callArgs
}

func (f mockFunc) writeOutArgs(w *writer) {
	if len(f.outs) == 1 {
		w.write(" " + f.outs[0].fullName())
	} else if len(f.outs) > 1 {
		w.write(" (")
		for i, a := range f.outs {
			if i > 0 {
				w.write(", ")
			}
			w.write(a.fullName())
		}
		w.write(")")
	}
}

const (
	returnVarName   = "retArgs"
	returnVarAssign = "\t" + returnVarName + " := "
)

func (f mockFunc) writeBody(callArgs []string, w *writer) {
	w.writeLines(" {")
	hasReturns := len(f.outs) > 0
	prefix := "\t"
	if hasReturns {
		prefix = returnVarAssign
	}
	callLines := make([]string, 0)
	if f.isVaradic {
		callLines = append(callLines, "\targs := make([]any, 0)")
		if len(callArgs) > 1 {
			callLines = append(callLines, "\targs = append(args, "+strings.Join(callArgs[:len(callArgs)-1], ", ")+")")
		}
		if f.ins[len(f.ins)-1].isAny() {
			callLines = append(callLines, "\targs = append(args, "+callArgs[len(callArgs)-1]+"...)")
		} else {
			callLines = append(callLines,
				"\tfor _, v := range "+callArgs[len(callArgs)-1]+" {",
				"\t\targs = append(args, v)",
				"\t}")
		}
		callLines = append(callLines, prefix+"m.Called(args...)")
	} else if len(f.ins) > 0 {
		callLines = append(callLines, prefix+"m.Called("+strings.Join(callArgs, ", ")+")")
	} else {
		callLines = append(callLines, prefix+"m.Called()")
	}
	w.writeLines(callLines...)
	if hasReturns {
		f.writeReturn(w)
	}
	w.writeLines("}")
}

func (f mockFunc) writeReturn(w *writer) {
	w.write("\treturn ")
	if l := len(f.outs); l <= 4 {
		oTypes := make([]string, l)
		for i, a := range f.outs {
			oTypes[i] = a.fullName()
		}
		w.write(fmt.Sprintf("mmock.As%d[%s](%s)", l, strings.Join(oTypes, ", "), returnVarName))
	} else {
		for i, a := range f.outs {
			if i > 0 {
				w.write(", ")
			}
			w.write(fmt.Sprintf("mmock.As[%s](%s, %d)", a.fullName(), returnVarName, i))
		}
	}
	w.writeLines("")
}

type mockArg struct {
	defPkg    string
	name      string
	pkg       string
	isPtr     bool
	isSlice   bool
	isMap     bool
	mapKey    *mockArg
	itemPfx   string
	isVaradic bool
}

func newMockArg(pkg string, arg reflect.Type, isVaradic bool) mockArg {
	ma := &mockArg{
		defPkg:    pkg,
		isVaradic: isVaradic,
	}
	itemPfx := ""
	if ma.isPtr = arg.Kind() == reflect.Pointer; ma.isPtr {
		arg = arg.Elem()
	}
	if ma.isSlice = arg.Kind() == reflect.Slice; ma.isSlice {
		arg, itemPfx = ma.finalType(arg.Elem(), "")
		if isVaradic {
			ma.isSlice = false
		}
	} else if ma.isMap = arg.Kind() == reflect.Map; ma.isMap {
		mk := newMockArg(pkg, arg.Key(), false)
		ma.mapKey = &mk
		arg, itemPfx = ma.finalType(arg.Elem(), "")
	}
	ma.name = arg.Name()
	ma.pkg = arg.PkgPath()
	ma.itemPfx = itemPfx
	return *ma
}

func (a mockArg) isAny() bool {
	return a.name == "" && !a.isPtr
}

func (a mockArg) finalType(rt reflect.Type, pfx string) (reflect.Type, string) {
	if rt.Kind() == reflect.Pointer {
		return a.finalType(rt.Elem(), pfx+"*")
	} else if rt.Kind() == reflect.Slice {
		return a.finalType(rt.Elem(), pfx+"[]")
	} else if rt.Kind() == reflect.Map {
		mk := newMockArg(a.defPkg, rt.Key(), false)
		return a.finalType(rt.Elem(), pfx+"map["+mk.fullName()+"]")
	}
	return rt, pfx
}

func (a mockArg) fullName() string {
	var builder strings.Builder
	if a.isPtr {
		builder.WriteString("*")
	}
	if a.isVaradic {
		builder.WriteString("...")
		builder.WriteString(a.itemPfx)
	} else if a.isSlice {
		builder.WriteString("[]")
		builder.WriteString(a.itemPfx)
	} else if a.isMap {
		builder.WriteString("map[")
		builder.WriteString(a.mapKey.fullName())
		builder.WriteString("]")
		builder.WriteString(a.itemPfx)
	}
	if pkg, _ := packagePathToPackage(a.pkg); pkg != "" && pkg != a.defPkg {
		builder.WriteString(pkg + ".")
	}
	if a.name != "" {
		builder.WriteString(a.name)
	} else {
		builder.WriteString("any")
	}
	return builder.String()
}

func packagePathToPackage(pp string) (string, bool) {
	if pp == "" {
		return pp, true
	}
	pts := strings.Split(pp, "/")
	if len(pts) == 1 {
		return pts[0], true
	}
	pkg := pts[len(pts)-1]
	if strings.HasPrefix(pkg, "v") && len(pkg) > 1 && pkg[1] >= '0' && pkg[1] <= '9' {
		pkg = pts[len(pts)-2]
	}
	return pkg, false
}

type writer struct {
	buffer bytes.Buffer
	w      *bufio.Writer
	err    error
}

func newWriter(w *bufio.Writer) *writer {
	res := &writer{}
	if w != nil {
		res.w = w
	} else {
		res.w = bufio.NewWriter(&res.buffer)
	}
	return res
}

func (w *writer) flush() error {
	if w.err == nil {
		w.err = w.w.Flush()
	}
	return w.err
}

func (w *writer) bytes() ([]byte, error) {
	_ = w.flush()
	return w.buffer.Bytes(), w.err
}

func (w *writer) write(s string) *writer {
	_, w.err = w.w.WriteString(s)
	return w
}

func (w *writer) writeLines(lns ...string) *writer {
	for _, ln := range lns {
		_, w.err = w.w.WriteString(ln + "\n")
	}
	return w
}
