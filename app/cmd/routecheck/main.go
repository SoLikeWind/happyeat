// routecheck 从 routes.go 解析挂 Casbin 的路由，与 casbinrules.PermissionRules 聚合的 (obj,act) 做 diff。
// 用法：在模块根目录 go run ./app/cmd/routecheck  [-json] [-strict] [-routes path]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/solikewind/happyeat/app/internal/pkg/casbinrules"
	"github.com/solikewind/happyeat/app/internal/pkg/routenorm"
)

type routeHit struct {
	Method   string `json:"method"`
	Path     string `json:"path"`
	CanonKey string `json:"key"`
}

type zombieHit struct {
	Method   string `json:"method"`
	Obj      string `json:"obj"`
	CanonKey string `json:"key"`
}

func main() {
	jsonOut := flag.Bool("json", false, "JSON 输出（供 CI）")
	strict := flag.Bool("strict", false, "僵尸规则也视为失败（非零退出码）")
	routesPath := flag.String("routes", filepath.Join("app", "internal", "handler", "routes.go"), "routes.go 路径（相对当前工作目录）")
	flag.Parse()

	abs, err := filepath.Abs(*routesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "routecheck: %v\n", err)
		os.Exit(2)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, abs, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "routecheck: 解析 %s: %v\n", abs, err)
		os.Exit(2)
	}

	routes, err := extractCasbinRoutes(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "routecheck: %v\n", err)
		os.Exit(2)
	}

	routeKeys := make(map[string]routeHit)
	for _, r := range routes {
		// 自助接口由 Casbin 中间件放行（仅需登录），不纳入 permission 规则一致性校验。
		if routenorm.IsSelfServicePath(r.FullPath) {
			continue
		}
		k := policyKey(r.FullPath, r.Method)
		routeKeys[k] = routeHit{Method: r.Method, Path: r.FullPath, CanonKey: k}
	}

	policyKeys := casbinrules.AllPolicyKeys(policyKey)

	var missing []routeHit
	for k, h := range routeKeys {
		if _, ok := policyKeys[k]; !ok {
			missing = append(missing, h)
		}
	}
	sort.Slice(missing, func(i, j int) bool {
		if missing[i].Path == missing[j].Path {
			return missing[i].Method < missing[j].Method
		}
		return missing[i].Path < missing[j].Path
	})

	var zombies []zombieHit
	for k := range policyKeys {
		if _, ok := routeKeys[k]; !ok {
			obj, act := splitPolicyKey(k)
			zombies = append(zombies, zombieHit{Method: act, Obj: obj, CanonKey: k})
		}
	}
	sort.Slice(zombies, func(i, j int) bool {
		if zombies[i].Obj == zombies[j].Obj {
			return zombies[i].Method < zombies[j].Method
		}
		return zombies[i].Obj < zombies[j].Obj
	})

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{
			"missing_in_permissions": missing,
			"zombie_rules":           zombies,
		})
	} else {
		if len(missing) > 0 {
			fmt.Println("未出现在任何 permission 规则中（高风险漏拒绝）：")
			for _, m := range missing {
				fmt.Printf("  %s %s  [%s]\n", m.Method, m.Path, m.CanonKey)
			}
		} else {
			fmt.Println("未出现在任何 permission 规则中：无")
		}
		fmt.Println()
		if len(zombies) > 0 {
			fmt.Println("僵尸规则（permission 有、当前 Casbin 路由无）：")
			for _, z := range zombies {
				fmt.Printf("  %s %s  [%s]\n", z.Method, z.Obj, z.CanonKey)
			}
		} else {
			fmt.Println("僵尸规则：无")
		}
	}

	exit := 0
	if len(missing) > 0 {
		exit = 1
	}
	if *strict && len(zombies) > 0 {
		exit = 1
	}
	os.Exit(exit)
}

func policyKey(obj, act string) string {
	return routenorm.CanonicalObj(obj) + "|" + strings.ToUpper(strings.TrimSpace(act))
}

func splitPolicyKey(k string) (obj, act string) {
	i := strings.LastIndex(k, "|")
	if i < 0 {
		return k, ""
	}
	return k[:i], k[i+1:]
}

type parsedRoute struct {
	Method   string
	FullPath string
}

func extractCasbinRoutes(file *ast.File) ([]parsedRoute, error) {
	var fn *ast.FuncDecl
	for _, d := range file.Decls {
		f, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if f.Name != nil && f.Name.Name == "RegisterHandlers" {
			fn = f
			break
		}
	}
	if fn == nil || fn.Body == nil {
		return nil, fmt.Errorf("未找到 RegisterHandlers")
	}

	var out []parsedRoute
	for _, stmt := range fn.Body.List {
		es, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}
		call, ok := es.X.(*ast.CallExpr)
		if !ok || !isAddRoutesCall(call) {
			continue
		}
		prefix := findWithPrefix(call.Args)
		mw := findWithMiddlewares(call.Args)
		if mw == nil || len(mw.Args) < 2 {
			continue
		}
		if !middlewareSliceHasCasbin(mw.Args[0]) {
			continue
		}
		routeLit := unwrapRouteComposite(mw.Args[1])
		if routeLit == nil {
			continue
		}
		for _, elt := range routeLit.Elts {
			cl, ok := elt.(*ast.CompositeLit)
			if !ok {
				continue
			}
			method, path := extractMethodPath(cl)
			if method == "" || path == "" {
				continue
			}
			full := joinURLPath(prefix, path)
			out = append(out, parsedRoute{Method: method, FullPath: full})
		}
	}
	return out, nil
}

func isAddRoutesCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil {
		return false
	}
	return sel.Sel.Name == "AddRoutes"
}

func findWithPrefix(args []ast.Expr) string {
	for _, a := range args {
		ce := unwrapParensCall(a)
		if ce == nil {
			continue
		}
		if !isRestIdent(ce.Fun, "WithPrefix") || len(ce.Args) < 1 {
			continue
		}
		if s, ok := stringFromExpr(ce.Args[0]); ok {
			return s
		}
	}
	return ""
}

func findWithMiddlewares(args []ast.Expr) *ast.CallExpr {
	for _, a := range args {
		ce := unwrapParensCall(a)
		if ce == nil {
			continue
		}
		if isRestIdent(ce.Fun, "WithMiddlewares") {
			return ce
		}
	}
	return nil
}

func unwrapParensCall(e ast.Expr) *ast.CallExpr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			break
		}
		e = p.X
	}
	ce, ok := e.(*ast.CallExpr)
	if !ok {
		return nil
	}
	return ce
}

func isRestIdent(fun ast.Expr, name string) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != name {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "rest"
}

func middlewareSliceHasCasbin(e ast.Expr) bool {
	cl, ok := e.(*ast.CompositeLit)
	if !ok {
		return false
	}
	for _, elt := range cl.Elts {
		if isCasbinMiddlewareRef(elt) {
			return true
		}
	}
	return false
}

func isCasbinMiddlewareRef(e ast.Expr) bool {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "CasbinMiddleware" {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "serverCtx"
}

func unwrapRouteComposite(e ast.Expr) *ast.CompositeLit {
	if u, ok := e.(*ast.UnaryExpr); ok && u.Op == token.ELLIPSIS {
		e = u.X
	}
	cl, ok := e.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	return cl
}

func extractMethodPath(cl *ast.CompositeLit) (method, path string) {
	for _, f := range cl.Elts {
		kv, ok := f.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		id, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch id.Name {
		case "Method":
			method = methodFromExpr(kv.Value)
		case "Path":
			if s, ok := stringFromExpr(kv.Value); ok {
				path = s
			}
		}
	}
	return method, path
}

func methodFromExpr(e ast.Expr) string {
	if s, ok := stringFromExpr(e); ok {
		return strings.ToUpper(s)
	}
	sel, ok := e.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil {
		return ""
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok || id.Name != "http" {
		return ""
	}
	switch sel.Sel.Name {
	case "MethodGet":
		return "GET"
	case "MethodPost":
		return "POST"
	case "MethodPut":
		return "PUT"
	case "MethodDelete":
		return "DELETE"
	case "MethodPatch":
		return "PATCH"
	case "MethodHead":
		return "HEAD"
	case "MethodOptions":
		return "OPTIONS"
	default:
		return ""
	}
}

func stringFromExpr(e ast.Expr) (string, bool) {
	bl, ok := e.(*ast.BasicLit)
	if !ok || bl.Kind != token.STRING {
		return "", false
	}
	return strings.Trim(bl.Value, `"`), true
}

func joinURLPath(prefix, p string) string {
	prefix = strings.TrimRight(prefix, "/")
	if p == "" {
		return prefix
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if prefix == "" {
		return strings.TrimRight(p, "/")
	}
	return strings.TrimRight(prefix+p, "/")
}
