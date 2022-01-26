package util

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"

	"gitee.com/Rainkropy/frm-lib/util"
)

// 文件
type AstFile struct {
	// 包名
	PackageName string
	// 导入的包
	Imports map[string]*AstImport
	// 方法
	Funcs []AstFunc
}

// 结构体
type AstStruct struct {
	PackageName string           // 包名称
	Name        string           // 结构体名称
	Doc         string           // 结构体注释说明，使用@name注解
	Field       *AstStructField  // 如果是类型别名，则包含该字段，比如 type Mystring string 这种格式的结构体
	Fields      []AstStructField // 包含的字段
}

// 结构体字段
type AstStructField struct {
	Name        string   // 字段名称
	Doc         string   // 注释说明，默认第一行
	Type        string   // 类型
	PackageName string   // 所属的包，如果类型不是常用类型，则通过包名和类型来递归解析
	Tags        []string // 标签，使用@tag注解
	JsType      string   // 使用的js类型，使用@jsType注解
}

// 方法
type AstFunc struct {
	// 方法名称
	FuncName string
	// 接口名称
	ApiName string
	// 请求方法
	Method string
	// 使用的中间件
	MiddleWares []string
	// 请求地址
	URL string
	// 作用域
	Scope string
	// 自定义参数
	Requests map[string]string
	// 参数, 参数名: 参数类型
	Params []AstParam
	// 注释
	Comments map[string][]string
	// 返回值
	Results []AstResult
}

// 引用
type AstImport struct {
	Name  string
	Value string
	Used  bool
}

// 参数
type AstParam struct {
	Name       string
	Type       string
	JsonName   string
	StructName string
	Custom     string
	Doc        string
}

// 返回值
type AstResult struct {
	Name    string
	Type    string
	Defined bool
}

const (
	// 用于定义API名称，如：@api 用户列表
	AST_API_NAME = "api"
	// 用于定义GET请求的URL，如：@get /users/api/list
	AST_API_GET = "get"
	// 用于定义POST请求的URL，如：@post /users/api/save
	AST_API_POST = "post"
	// 用于定义中间件，如：@mid VerifyUser VerifyIP
	AST_API_MID = "mid"
	// 用于自定义参数，如：@request ip app.GetIP()
	AST_API_REQ = "request"
	// 用于参数说明，如：@params id 用户ID
	AST_API_PARAMS = "params"

	// 名称注解
	AST_NAME = "name"
	// 标签注解
	AST_TAG = "tag"
	// JS类型注解
	AST_JSTYPE = "jsType"
)

func (f *AstFunc) GetRequestStr() string {
	if len(f.Params) < 1 && f.Scope == "" {
		return ""
	}
	req := "req := struct{\n"
	if f.Scope != "" {
		req += "\t" + f.Scope + "\n"
	}
	for _, p := range f.Params {
		req += fmt.Sprintf("\t%s %s `form:\"%s\"`\n", p.StructName, p.Type, p.JsonName)
	}
	req += "\t}\n"
	return req
}

// 获取AST
func GetAst(file string) (*token.FileSet, *ast.File, error) {
	reader, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, nil, err
	}
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", string(reader), parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	return fset, f, nil
}

// 获取文件中引入的包
func getImports(f *ast.File) map[string]*AstImport {
	// 导入的包名
	imports := make(map[string]*AstImport)
	for _, decl := range f.Decls {
		imp, ok := decl.(*ast.GenDecl)
		if !ok || imp.Tok != token.IMPORT {
			continue
		}
		for _, spec := range imp.Specs {
			if isp, ok := spec.(*ast.ImportSpec); ok {
				impName := ""
				if isp.Name != nil {
					impName = isp.Name.Name
				} else {
					path := strings.Trim(isp.Path.Value, "\"")
					ns := strings.Split(path, "/")
					impName = ns[len(ns)-1]
				}
				imports[impName] = &AstImport{
					Name:  impName,
					Value: isp.Path.Value,
					Used:  false,
				}
			}
		}
	}
	return imports
}

// 解析路由文件
func ParseRouterFile(file string) (*AstFile, error) {
	_, f, err := GetAst(file)
	if err != nil {
		return nil, err
	}
	// 包名
	pkg := f.Name.Name

	// 导入的包名
	imports := getImports(f)

	funcs := make([]AstFunc, 0)

	for _, decl := range f.Decls {
		fun, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// 方法没有注释，则返回
		if fun.Doc == nil {
			continue
		}

		astFunc := AstFunc{
			FuncName: fun.Name.Name,
		}

		requests := make(map[string]string)
		paramsDoc := make(map[string]string)
		params := make([]AstParam, 0)
		comments := make(map[string][]string)
		results := make([]AstResult, 0)

		// 解析注释
		for _, comment := range fun.Doc.List {
			k, v := parseAstComment(comment.Text)
			switch k {
			case AST_API_NAME:
				astFunc.ApiName = v[0]
			case AST_API_MID:
				astFunc.MiddleWares = v
			case AST_API_GET:
				astFunc.Method = "GET"
				astFunc.URL = v[0]
			case AST_API_POST:
				astFunc.Method = "POST"
				astFunc.URL = v[0]
			case AST_API_REQ:
				if len(v) == 2 {
					requests[v[0]] = v[1]
				}
			case AST_API_PARAMS:
				if len(v) > 1 {
					paramsDoc[v[0]] = util.Implode(" ", v[1:])
				}
			default:
				if ev, ok := comments[k]; ok {
					ev = append(ev, v...)
					comments[k] = ev
				} else {
					comments[k] = v
				}
			}
		}

		if astFunc.ApiName == "" || astFunc.Method == "" || astFunc.URL == "" {
			continue
		}

		// 解析前缀
		if fun.Recv != nil && len(fun.Recv.List) > 0 {
			recv := fun.Recv.List[0]
			if expr, ok := recv.Type.(*ast.StarExpr); ok {
				if ident, ok2 := expr.X.(*ast.Ident); ok2 {
					astFunc.Scope = ident.Name
				}
			}
		}

		// 解析参数
		if fun.Type != nil && fun.Type.Params != nil && len(fun.Type.Params.List) > 0 {
			for _, field := range fun.Type.Params.List {
				typeName := parseAstExprName(field.Type)
				if pos := util.Strpos(typeName, ".", 0); pos > 0 {
					impName := util.Substr(typeName, 0, pos)
					if n, ok := imports[impName]; ok {
						n.Used = true
					}
				}
				for _, fieldName := range field.Names {
					doc := ""
					if d, ok := paramsDoc[fieldName.Name]; ok {
						doc = d
					}
					custom := ""
					if c, ok := requests[fieldName.Name]; ok {
						custom = c
					}
					params = append(params, AstParam{
						Name:       fieldName.Name,
						Type:       typeName,
						StructName: Ucfirst(Camel(fieldName.Name)),
						JsonName:   Snake(fieldName.Name),
						Custom:     custom,
						Doc:        doc,
					})
				}
			}
		}

		// 解析返回值
		if fun.Type != nil && fun.Type.Results != nil && len(fun.Type.Results.List) > 0 {
			for idx, result := range fun.Type.Results.List {
				typeName := parseAstExprName(result.Type)
				if result.Names != nil && len(result.Names) > 0 {
					for _, fieldName := range result.Names {
						results = append(results, AstResult{
							Name:    fieldName.Name,
							Type:    typeName,
							Defined: true,
						})
					}
				} else {
					results = append(results, AstResult{
						Name:    fmt.Sprintf("_R%d", idx),
						Type:    typeName,
						Defined: false,
					})
				}

			}
		}
		astFunc.Requests = requests
		astFunc.Params = params
		astFunc.Comments = comments
		astFunc.Results = results

		funcs = append(funcs, astFunc)

	}

	rest := AstFile{
		PackageName: pkg,
		Imports:     imports,
		Funcs:       funcs,
	}

	return &rest, nil
}

// 解析类型
func parseAstExprName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StarExpr:
		return parseAstExprName(v.X)
	case *ast.ArrayType:
		return "[]" + parseAstExprName(v.Elt)
	case *ast.MapType:
		mpkey := parseAstExprName(v.Key)
		mpval := parseAstExprName(v.Value)
		return fmt.Sprintf("map[%s]%s", mpkey, mpval)
	case *ast.SelectorExpr:
		pkg := parseAstExprName(v.X)
		return fmt.Sprintf("%s.%s", pkg, v.Sel.Name)
	default:
		return ""
	}
	return ""
}

// 解析注释
func parseAstComment(commentText string) (string, []string) {
	key := ""
	s := ""
	if Substr(commentText, 0, 3) == "//@" {
		s = Substr(commentText, 3, len(commentText)-2)
	} else if Substr(commentText, 0, 4) == "// @" {
		s = Substr(commentText, 4, len(commentText)-3)
	} else {
		return "", nil
	}
	ss := Explode(" ", strings.Trim(s, " "))
	value := make([]string, 0)
	for idx, sd := range ss {
		if idx == 0 {
			key = sd
		} else if sd != "" {
			value = append(value, sd)
		}
	}
	return key, value
}
