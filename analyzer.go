package canonicalheader

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"net/http"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

const (
	pkgPath = "net/http"
	name    = "Header"
)

var errInvalidInspectResult = errors.New("unexpected inspect analyzer result type")

//nolint:gochecknoglobals // for backward compatibility.
var Analyzer = New()

func New() *analysis.Analyzer {
	c := &canonicalHeader{}

	a := &analysis.Analyzer{
		Name:     "canonicalheader",
		Doc:      "canonicalheader checks whether net/http.Header uses canonical header",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      c.run,
	}

	a.Flags.Init("canonicalheader", flag.ExitOnError)
	a.Flags.Var(&c.exclusions, "exclusions", "comma-separated list of exclusion rules")
	a.Flags.BoolVar(&c.useDefaultExclusion, "useDefaultExclusion", true, "use default exclusion rules")

	return a
}

type canonicalHeader struct {
	useDefaultExclusion bool
	exclusions          stringSet
}

type argumenter interface {
	diagnostic(canonicalHeader string) analysis.Diagnostic
	value() string
}

func (c *canonicalHeader) buildExclusions() map[string]string {
	var defEx map[string]string
	if c.useDefaultExclusion {
		defEx = initialism()
	}

	result := make(map[string]string, len(c.exclusions)+len(defEx))
	maps.Copy(result, defEx)

	for _, ex := range c.exclusions {
		result[http.CanonicalHeaderKey(ex)] = ex
	}

	return result
}

func (c *canonicalHeader) run(pass *analysis.Pass) (any, error) {
	headerObject := findHeaderObject(pass)
	if headerObject == nil {
		//nolint:nilnil // Analyzer has nil ResultType, so nil result is expected.
		return nil, nil
	}

	spctor, err := inspectorFromPass(pass)
	if err != nil {
		return nil, err
	}

	exclusions := c.buildExclusions()

	spctor.WithStack([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}

		callExp, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		reportCall(pass, callExp, reportCallContext{
			stack:      stack,
			headerType: headerObject.Type(),
			exclusions: exclusions,
		})
		return true
	})

	//nolint:nilnil // Analyzer has nil ResultType, so nil result is expected.
	return nil, nil
}

func findHeaderObject(pass *analysis.Pass) types.Object {
	if pass == nil || pass.Pkg == nil {
		return nil
	}

	for _, importedPkg := range pass.Pkg.Imports() {
		if importedPkg == nil || importedPkg.Path() != pkgPath {
			continue
		}

		return importedPkg.Scope().Lookup(name)
	}

	return nil
}

func inspectorFromPass(pass *analysis.Pass) (*inspector.Inspector, error) {
	spctor, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, fmt.Errorf("%w: want %T, got %T", errInvalidInspectResult, spctor, pass.ResultOf[inspect.Analyzer])
	}

	return spctor, nil
}

func reportCall(
	pass *analysis.Pass,
	callExp *ast.CallExpr,
	ctx reportCallContext,
) {
	callInfo, ok := resolveReceiverAndMethod(pass, callExp, ctx.stack)
	if !ok || !types.Identical(callInfo.recvType, ctx.headerType) || !isValidMethod(callInfo.methodName) {
		return
	}

	arg, ok := resolveCallArgument(pass.TypesInfo, callExp)
	if !ok {
		return
	}

	argValue := arg.value()
	headerKeyCanonical := canonicalHeaderKey(argValue, ctx.exclusions)
	if argValue == headerKeyCanonical {
		return
	}

	pass.Report(arg.diagnostic(headerKeyCanonical))
}

type callInfo struct {
	methodName string
	recvType   types.Type
}

type reportCallContext struct {
	stack      []ast.Node
	headerType types.Type
	exclusions map[string]string
}

func resolveReceiverAndMethod(pass *analysis.Pass, callExp *ast.CallExpr, stack []ast.Node) (callInfo, bool) {
	switch callee := typeutil.Callee(pass.TypesInfo, callExp).(type) {
	case *types.Func:
		return resolveDirectMethodCall(callee, callExp)
	case *types.Var:
		resolvedMethod, ok := resolveMethodValueCall(pass, callExp, stack)
		if !ok {
			return callInfo{}, false
		}

		return callInfo(resolvedMethod), true
	default:
		return callInfo{}, false
	}
}

func resolveDirectMethodCall(fn *types.Func, callExp *ast.CallExpr) (callInfo, bool) {
	signature, ok := fn.Type().(*types.Signature)
	if !ok {
		return callInfo{}, false
	}

	recv := signature.Recv()
	if recv == nil {
		return callInfo{}, false
	}

	sel, ok := callExp.Fun.(*ast.SelectorExpr)
	if !ok {
		return callInfo{}, false
	}

	return callInfo{
		methodName: sel.Sel.Name,
		recvType:   recv.Type(),
	}, true
}

func resolveCallArgument(info *types.Info, callExp *ast.CallExpr) (argumenter, bool) {
	if len(callExp.Args) == 0 {
		return nil, false
	}

	callArg, ok := unwrapTypeCastArg(info, callExp.Args[0])
	if !ok {
		return nil, false
	}

	switch t := callArg.(type) {
	case *ast.BasicLit:
		lString, err := newLiteralString(t)
		if err != nil {
			return nil, false
		}

		return lString, true
	case *ast.Ident:
		constString, err := newConstantKey(info, t)
		if err != nil {
			return nil, false
		}

		return constString, true
	default:
		return nil, false
	}
}

func unwrapTypeCastArg(info *types.Info, expr ast.Expr) (ast.Expr, bool) {
	callArg := expr

	for {
		// If it is not *ast.CallExpr, this is a value.
		call, ok := callArg.(*ast.CallExpr)
		if !ok {
			return callArg, true
		}

		// Some function is called, skip this case.
		if len(call.Args) == 0 {
			return nil, false
		}

		f, ok := call.Fun.(*ast.Ident)
		if !ok {
			return callArg, true
		}

		obj := info.ObjectOf(f)
		// nil may be by code, but not by logic.
		// TypeInfo should contain of type.
		if obj == nil {
			return callArg, true
		}

		// This is function.
		// Skip this method call.
		_, ok = obj.Type().(*types.Signature)
		if ok {
			return nil, false
		}

		callArg = call.Args[0]
	}
}

type methodValueState struct {
	methodName    string
	recvType      types.Type
	hasValue      bool
	isMethodValue bool
}

type methodValue struct {
	methodName string
	recvType   types.Type
}

type methodValueUpdate struct {
	object types.Object
	state  methodValueState
}

type methodValueCallContext struct {
	callPos      token.Pos
	targetObject types.Object
	funcBodies   []*ast.BlockStmt
}

func resolveMethodValueCall(pass *analysis.Pass, callExp *ast.CallExpr, stack []ast.Node) (methodValue, bool) {
	ctx, ok := newMethodValueCallContext(pass, callExp, stack)
	if !ok {
		return methodValue{}, false
	}

	aliasStates := make(map[types.Object]methodValueState)
	for _, funcBody := range ctx.funcBodies {
		updateMethodValueStateByBody(pass, funcBody, ctx.callPos, aliasStates)
	}

	resolvedState, ok := aliasStates[ctx.targetObject]
	if !ok || !resolvedState.hasValue || !resolvedState.isMethodValue {
		return methodValue{}, false
	}

	return methodValue{
		methodName: resolvedState.methodName,
		recvType:   resolvedState.recvType,
	}, true
}

func newMethodValueCallContext(pass *analysis.Pass, callExp *ast.CallExpr, stack []ast.Node) (methodValueCallContext, bool) {
	if pass == nil || pass.TypesInfo == nil || callExp == nil || len(stack) == 0 {
		return methodValueCallContext{}, false
	}

	ident, ok := callExp.Fun.(*ast.Ident)
	if !ok {
		return methodValueCallContext{}, false
	}

	targetObject := pass.TypesInfo.ObjectOf(ident)
	if targetObject == nil {
		return methodValueCallContext{}, false
	}

	funcBodies := enclosingFunctionBodies(stack)
	if len(funcBodies) == 0 {
		return methodValueCallContext{}, false
	}

	return methodValueCallContext{
		callPos:      callExp.Pos(),
		targetObject: targetObject,
		funcBodies:   funcBodies,
	}, true
}

func enclosingFunctionBodies(path []ast.Node) []*ast.BlockStmt {
	bodies := make([]*ast.BlockStmt, 0, 2)

	for _, node := range path {
		switch n := node.(type) {
		case *ast.FuncDecl:
			if n.Body != nil {
				bodies = append(bodies, n.Body)
			}
		case *ast.FuncLit:
			if n.Body != nil {
				bodies = append(bodies, n.Body)
			}
		}
	}

	return bodies
}

func updateMethodValueStateByBody(
	pass *analysis.Pass,
	body *ast.BlockStmt,
	callPos token.Pos,
	aliasStates map[types.Object]methodValueState,
) {
	if body == nil {
		return
	}

	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil {
			return true
		}

		if node.Pos() > callPos {
			return false
		}

		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}

		switch n := node.(type) {
		case *ast.AssignStmt:
			updateMethodValueStateFromAssign(pass, n, aliasStates)
		case *ast.ValueSpec:
			updateMethodValueStateFromValueSpec(pass, n, aliasStates)
		}

		return true
	})
}

func updateMethodValueStateFromAssign(
	pass *analysis.Pass,
	assignStmt *ast.AssignStmt,
	aliasStates map[types.Object]methodValueState,
) {
	updates := make([]methodValueUpdate, 0, len(assignStmt.Lhs))
	for idx, lhs := range assignStmt.Lhs {
		lhsIdent, ok := lhs.(*ast.Ident)
		if !ok {
			continue
		}

		lhsObject := pass.TypesInfo.ObjectOf(lhsIdent)
		if lhsObject == nil {
			continue
		}

		rhsExpr, ok := exprForIndex(assignStmt.Rhs, len(assignStmt.Lhs), idx)
		if !ok {
			updates = append(updates, methodValueUpdate{
				object: lhsObject,
				state: methodValueState{
					hasValue:      true,
					isMethodValue: false,
				},
			})
			continue
		}

		updates = append(updates, methodValueUpdate{
			object: lhsObject,
			state:  methodValueStateByExpr(pass, rhsExpr, aliasStates),
		})
	}

	for _, update := range updates {
		aliasStates[update.object] = update.state
	}
}

func updateMethodValueStateFromValueSpec(
	pass *analysis.Pass,
	valueSpec *ast.ValueSpec,
	aliasStates map[types.Object]methodValueState,
) {
	updates := make([]methodValueUpdate, 0, len(valueSpec.Names))
	for idx, name := range valueSpec.Names {
		lhsObject := pass.TypesInfo.ObjectOf(name)
		if lhsObject == nil {
			continue
		}

		rhsExpr, ok := exprForIndex(valueSpec.Values, len(valueSpec.Names), idx)
		if !ok {
			updates = append(updates, methodValueUpdate{
				object: lhsObject,
				state: methodValueState{
					hasValue:      true,
					isMethodValue: false,
				},
			})
			continue
		}

		updates = append(updates, methodValueUpdate{
			object: lhsObject,
			state:  methodValueStateByExpr(pass, rhsExpr, aliasStates),
		})
	}

	for _, update := range updates {
		aliasStates[update.object] = update.state
	}
}

func methodValueStateByExpr(
	pass *analysis.Pass,
	expr ast.Expr,
	aliasStates map[types.Object]methodValueState,
) methodValueState {
	methodValue, ok := methodValueByExpr(pass, expr)
	if !ok {
		ident, isIdent := ast.Unparen(expr).(*ast.Ident)
		if isIdent {
			sourceObject := pass.TypesInfo.ObjectOf(ident)
			sourceState, hasSource := aliasStates[sourceObject]
			if hasSource && sourceState.hasValue && sourceState.isMethodValue {
				return methodValueState{
					methodName:    sourceState.methodName,
					recvType:      sourceState.recvType,
					hasValue:      true,
					isMethodValue: true,
				}
			}
		}

		return methodValueState{
			hasValue:      true,
			isMethodValue: false,
		}
	}

	return methodValueState{
		methodName:    methodValue.methodName,
		recvType:      methodValue.recvType,
		hasValue:      true,
		isMethodValue: true,
	}
}

func methodValueByExpr(pass *analysis.Pass, expr ast.Expr) (methodValue, bool) {
	if pass == nil || pass.TypesInfo == nil || expr == nil {
		return methodValue{}, false
	}

	sel, ok := ast.Unparen(expr).(*ast.SelectorExpr)
	if !ok {
		return methodValue{}, false
	}

	selection := pass.TypesInfo.Selections[sel]
	if selection == nil || selection.Kind() != types.MethodVal {
		return methodValue{}, false
	}

	return methodValue{
		methodName: sel.Sel.Name,
		recvType:   selection.Recv(),
	}, true
}

func exprForIndex(expressions []ast.Expr, lhsCount, index int) (ast.Expr, bool) {
	if index < 0 || len(expressions) == 0 {
		return nil, false
	}

	if lhsCount == len(expressions) && index < len(expressions) {
		return expressions[index], true
	}

	if lhsCount == 1 && len(expressions) == 1 && index == 0 {
		return expressions[0], true
	}

	return nil, false
}

func canonicalHeaderKey(s string, m map[string]string) string {
	canonical := http.CanonicalHeaderKey(s)

	wellKnown, ok := m[canonical]
	if !ok {
		return canonical
	}

	return wellKnown
}

func isValidMethod(name string) bool {
	switch name {
	case "Get", "Set", "Add", "Del", "Values":
		return true
	default:
		return false
	}
}

func typeOfIdent(info *types.Info, ident *ast.Ident) (types.Type, bool) {
	if info == nil || ident == nil {
		return nil, false
	}

	obj := info.ObjectOf(ident)
	if obj == nil {
		return nil, false
	}

	return obj.Type(), true
}
