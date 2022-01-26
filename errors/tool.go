package errors

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

// 解析错误文件
// @param string filePath    错误定义文件
// @param string toFilePath  生成的字典文件
// @param string packageName 生成字典文件的包名
func ParseErrors(filePath, toFilePath, packageName string) map[int]string {
	errorMap := make(map[int]string)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		panic(err)
	}

	reader, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", string(reader), parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for _, i := range f.Decls {
		decl, ok := i.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, s := range decl.Specs {
			var code int
			var msg string
			if spec, ok := s.(*ast.ValueSpec); ok {
				// 判断是否为const常量
				if spec.Names[0].Obj.Kind != ast.Con {
					continue
				}

				// 获取常量的值并转换为int值
				if lit, ok2 := spec.Values[0].(*ast.BasicLit); ok2 {
					if v, err := strconv.Atoi(lit.Value); err == nil {
						code = v
					}
				}

				//获取注释
				if spec.Doc != nil {
					msg = strings.TrimSpace(spec.Doc.List[0].Text[2:])
				} else if decl.Doc != nil {
					msg = strings.TrimSpace(decl.Doc.List[0].Text[2:])
				} else {
					msg = "未知错误"
				}
			}

			errorMap[code] = msg
		}
	}
	errorFileStr := toErrorsGoFileStr(errorMap, packageName)
	writeStrToFile(errorFileStr, toFilePath)
	return errorMap
}

// 将错误map生成为文件字符
func toErrorsGoFileStr(m map[int]string, packageName string) string {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	keys = keysSort(keys)

	str := "package " + packageName + "\n\n"
	str += "var Errors = map[int]string{\n"
	for _, k := range keys {
		lineStr := fmt.Sprintf("\t%d : \"%s\",\n", k, m[k])
		str += lineStr
	}
	str += "}"
	return str
}

// 写入文件内容
func writeStrToFile(str string, filePath string) error {
	fp := path.Dir(filePath)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		os.Mkdir(fp, 0777)
		os.Chmod(fp, 0777)
	}
	var f *os.File
	var err error
	os.Remove(filePath)
	f, err = os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(str)
	return err
}

// 排序
func keysSort(data []int) []int {
	for i := 0; i < len(data); i++ {
		for j := i + 1; j < len(data); j++ {
			if data[i] > data[j] {
				data[j], data[i] = data[i], data[j]
			}
		}
	}
	return data
}

// 将错误码写入到markdown文件
func WriteErrorsToMD(mdFile string) error {
	keys := make([]int, 0, len(Errors))
	for k := range Errors {
		keys = append(keys, k)
	}
	keys = keysSort(keys)

	str := "## Error Code\n"
	str += "| code | msg |\n"
	str += "| -- | -- | \n"
	for _, k := range keys {
		str += fmt.Sprintf("| %d | %s |\n", k, Errors[k])
	}
	return writeStrToFile(str, mdFile)
}
