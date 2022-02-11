/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/cucumber/gherkin-go/v11"
	"github.com/cucumber/messages-go/v10"
	"github.com/iancoleman/orderedmap"
	"golang.org/x/tools/go/ast/astutil"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/bfenetworks/ingress-bfe/test/e2e/pkg/files"
)

var codeGenTemplate *template.Template

func main() {
	var (
		update            bool
		features          []string
		destPath          string
		generatorTemplate string
		testMainPath      string

		basePackage string
	)

	flag.BoolVar(&update, "update", false, "update files in place in case of missing steps or method definitions")
	flag.StringVar(&destPath, "dest-path", "test", "path to generated test package location")
	flag.StringVar(&generatorTemplate, "code-generator-template", "hack/codegen.tmpl", "path to the go template for code generation")
	flag.StringVar(&testMainPath, "test-main", "", "path to the TestMain go file")
	flag.StringVar(&basePackage, "base-package", "github.com/bfenetworks/ingress-bfe", "base go package")

	flag.Parse()

	// 1. verify flags
	features = flag.CommandLine.Args()
	if len(features) == 0 {
		fmt.Println("Usage: codegen [-update=false] [-dest-path=steps [features]")
		fmt.Println()
		fmt.Println("Example: codegen features/default_backend.feature")
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	// 2. parse template
	var err error
	codeGenTemplate, err = template.New("codegen.tmpl").Funcs(templateFuncs).ParseFiles(generatorTemplate)

	if err != nil {
		log.Fatalf("Unexpected error parsing template: %v", err)
	}

	// 3. if features is a directory, iterate and search for files with extension .feature
	if len(features) == 1 && files.IsDir(features[0]) {
		root := features[0]
		features = []string{}

		err := filepath.Walk(root, visitDir(&features))
		if err != nil {
			log.Fatalf("Unexpected error reading directory %v: %v", root, err)
		}
	}

	// 4. iterate feature files
	for _, path := range features {
		err := processFeature(path, destPath, update, basePackage, testMainPath)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 16. last step verifies the TestMain file
	//     uses all the defined features
	if testMainPath == "" {
		return
	}

	featuresInTestMain, err := extractFeaturesMapKeys(testMainPath)
	if err != nil {
		log.Fatal(err)
	}

	featuresInTestMainSet := sets.NewString(featuresInTestMain...)
	featuresSet := sets.NewString(features...)

	if !featuresInTestMainSet.Equal(featuresSet) {
		log.Printf(`Generated features mapping from .features files differ from the expected in TestMain file %v
expected	%v
generated	%v

`,
			testMainPath, features, featuresInTestMain)
	}
}

func processFeature(path, destPath string, update bool, basePackage, testMainPath string) error {
	// 5. parse feature file
	featureSteps, err := parseFeature(path)
	if err != nil {
		return fmt.Errorf("parsing feature file: %w", err)
	}

	// 6. generate package name to use
	packageName := generatePackage(path)

	// 7. check if go source file exists
	goFile := filepath.Join(destPath, packageName, "steps.go")
	isGoFileOk := files.Exists(goFile)

	mapping := &Mapping{
		Package:      packageName,
		FeatureFile:  path,
		Features:     featureSteps,
		NewFunctions: featureSteps,
		GoFile:       goFile,
	}

	// 8. Extract functions from go source code
	if isGoFileOk {
		goFunctions, err := extractFuncs(goFile)
		if err != nil {
			return fmt.Errorf("extracting go functions: %w", err)
		}

		mapping.GoDefinitions = goFunctions
	}

	if isGoFileOk {
		inFeatures := sets.NewString()
		inGo := sets.NewString()

		for _, feature := range mapping.Features {
			inFeatures.Insert(feature.Name)
		}

		for _, gofunc := range mapping.GoDefinitions {
			inGo.Insert(gofunc.Name)
		}

		mapping.NewFunctions = []Function{}

		if newFunctions := inFeatures.Difference(inGo); newFunctions.Len() > 0 {
			log.Printf("Feature file %v contains %v new function/s", mapping.FeatureFile, newFunctions.Len())

			var funcs []Function
			for _, f := range newFunctions.List() {
				for _, feature := range mapping.Features {
					if feature.Name == f {
						funcs = append(funcs, feature)
						break
					}
				}
			}

			mapping.NewFunctions = funcs
		}

		// 9. check signatures are ok
		signatureChanges := extractSignatureChanges(mapping)
		if len(signatureChanges) != 0 {
			var argBuf bytes.Buffer

			for _, sc := range signatureChanges {
				argBuf.WriteString(fmt.Sprintf(`
function %v
	have %v
	want %v
`, sc.Function, sc.Have, sc.Want))
			}

			return fmt.Errorf("source file %v has a different signature/s:\n %v", mapping.GoFile, argBuf.String())
		}
	}

	// 10. New go feature file
	if !isGoFileOk {
		log.Printf("Generating new go file %v...", mapping.GoFile)
		// 11. Feature to go source code
		err = generateGoFile(mapping)
		if err != nil {
			return err
		}

		featurePackage := filepath.Join(basePackage, destPath, packageName)

		// 12. update map variable in e2e_test
		if testMainPath != "" {
			err = updateFeatureMapVariable(mapping.FeatureFile, mapping.Package, featurePackage, testMainPath)
			if err != nil {
				return err
			}
		}

		return nil
	}

	if !update {
		if len(mapping.NewFunctions) != 0 {
			return fmt.Errorf("generated code %s exist but out of date, set argument -update=true if you need update file", mapping.GoFile)
		}

		return nil
	}

	// 13. if update is set
	log.Printf("Updating go file %v...", mapping.GoFile)
	return updateGoTestFile(mapping.GoFile, mapping.NewFunctions)
}

// Function holds the definition of a function in a go file or godog step
type Function struct {
	// Name
	Name string
	// Expr Regexp to use in godog Step definition
	Expr string
	// Args function arguments
	// k = name of the argument
	// v = type of the argument
	Args *orderedmap.OrderedMap
}

type Mapping struct {
	Package string

	FeatureFile string
	Features    []Function

	GoFile        string
	GoDefinitions []Function

	NewFunctions []Function
}

// SignatureChange holds information about the definition of a go function
type SignatureChange struct {
	Function string
	Have     string
	Want     string
}

var templateFuncs = template.FuncMap{
	"backticked": func(s string) string {
		return "`" + s + "`"
	},
	"unescape": func(s string) template.HTML {
		return template.HTML(s)
	},
	"argsFromMap": argsFromMap,
}

// parseFeature parses a godog feature file returning the unique
// steps definitions
func parseFeature(path string) ([]Function, error) {
	data, err := files.Read(path)
	if err != nil {
		return nil, err
	}

	gd, err := gherkin.ParseGherkinDocument(bytes.NewReader(data), (&messages.Incrementing{}).NewId)
	if err != nil {
		return nil, err
	}

	scenarios := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	funcs := []Function{}
	for _, s := range scenarios {
		funcs = parseSteps(s.Steps, funcs)
	}

	return funcs, nil
}

// extractFuncs reads a file containing go source code and returns
// the functions defined in the file.
func extractFuncs(filePath string) ([]Function, error) {
	if !strings.HasSuffix(filePath, ".go") {
		return nil, fmt.Errorf("only files with go extension are valid")
	}

	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var funcs []Function

	var printErr error
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		index := 0
		args := orderedmap.New()
		for _, p := range fn.Type.Params.List {
			var typeNameBuf bytes.Buffer

			err := printer.Fprint(&typeNameBuf, fset, p.Type)
			if err != nil {
				printErr = err
				return false
			}

			if len(p.Names) == 0 {
				argName := fmt.Sprintf("arg%d", index+1)
				args.Set(argName, typeNameBuf.String())

				index++
				continue
			}

			for _, ag := range p.Names {
				argName := ag.String()
				args.Set(argName, typeNameBuf.String())
				index++
			}
		}

		// Go functions do not have an expression
		funcs = append(funcs, Function{Name: fn.Name.Name, Args: args})

		return true
	})

	if printErr != nil {
		return nil, printErr
	}

	return funcs, nil
}

func updateGoTestFile(filePath string, newFuncs []Function) error {
	fileSet := token.NewFileSet()

	node, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	var featureFunc *ast.FuncDecl
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if fn.Name.Name == "InitializeScenario" {
			featureFunc = fn
		}

		return true
	})

	if featureFunc == nil {
		return fmt.Errorf("file %v does not contains a FeatureFunct function", filePath)
	}

	// Add new functions
	astf, err := toAstFunctions(newFuncs)
	if err != nil {
		return err
	}

	node.Decls = append(node.Decls, astf...)

	// Update steps in InitializeScenario
	astSteps, err := toContextStepsfuncs(newFuncs)
	if err != nil {
		return err
	}

	featureFunc.Body.List = append(astSteps, featureFunc.Body.List...)

	var buffer bytes.Buffer
	if err = format.Node(&buffer, fileSet, node); err != nil {
		return fmt.Errorf("error formatting file %v: %w", filePath, err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %v: %w", filePath, err)
	}

	return ioutil.WriteFile(filePath, buffer.Bytes(), fileInfo.Mode())
}

func toContextStepsfuncs(funcs []Function) ([]ast.Stmt, error) {
	astStepsTpl := `
package codegen
func InitializeScenario() { {{ range . }}
	ctx.Step({{ backticked .Expr | unescape }}, {{ .Name }}){{end}}
}
`
	astFile, err := astFromTemplate(astStepsTpl, funcs)
	if err != nil {
		return nil, err
	}

	f := astFile.Decls[0].(*ast.FuncDecl)

	return f.Body.List, nil
}

func toAstFunctions(funcs []Function) ([]ast.Decl, error) {
	astFuncTpl := `
package codegen
{{ range . }}func {{ .Name }}{{ argsFromMap .Args false }} error {
	return godog.ErrPending
}

{{end}}
`
	astFile, err := astFromTemplate(astFuncTpl, funcs)
	if err != nil {
		return nil, err
	}

	return astFile.Decls, nil
}

func astFromTemplate(astFuncTpl string, funcs []Function) (*ast.File, error) {
	buf := bytes.NewBuffer(make([]byte, 0))

	astFuncs, err := template.New("ast").Funcs(templateFuncs).Parse(astFuncTpl)
	if err != nil {
		return nil, err
	}

	err = astFuncs.Execute(buf, funcs)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, "src.go", buf.String(), parser.ParseComments)
	if err != nil {
		return nil, err
	}

	return astFile, nil
}

// generatePackage returns the name of the
// package to use using the feature filename
func generatePackage(filePath string) string {
	base := path.Base(filePath)
	base = strings.ToLower(base)
	base = strings.ReplaceAll(base, "_", "")
	base = strings.ReplaceAll(base, ".feature", "")

	return base
}

func argsFromMap(args *orderedmap.OrderedMap, onlyType bool) string {
	s := "("

	for _, k := range args.Keys() {
		v, ok := args.Get(k)
		if !ok {
			continue
		}

		if onlyType {
			s += fmt.Sprintf("%v, ", v)
		} else {
			s += fmt.Sprintf("%v %v, ", k, v)
		}
	}

	if len(args.Keys()) > 0 {
		s = s[0 : len(s)-2]
	}

	return s + ")"
}

func visitDir(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if filepath.Ext(path) != ".feature" {
			return nil
		}

		*files = append(*files, path)
		return nil
	}
}

const mapVariableName = "features"

// extractFeaturesMapKeys extracts the keys from the features map defined in
// the main test file defined in a variable:
// features = map[string]func(*godog.Suite){}
func extractFeaturesMapKeys(testPath string) ([]string, error) {
	fset := token.NewFileSet()

	fileAst, err := parser.ParseFile(fset, testPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, declarations := range fileAst.Decls {
		switch decl := declarations.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				spec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for _, fn := range spec.Names {
					if fn.Name != mapVariableName {
						continue
					}

					features := fn.Obj.Decl.(*ast.ValueSpec).Values
					elts := features[0].(*ast.CompositeLit).Elts

					featureNames := []string{}
					for _, elt := range elts {
						val := elt.(*ast.KeyValueExpr).Key.(*ast.BasicLit).Value
						s, err := strconv.Unquote(val)
						if err != nil {
							featureNames = append(featureNames, val)
							continue
						}

						featureNames = append(featureNames, s)
					}

					return featureNames, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("there is no features variable in file %v", testPath)
}

func updateFeatureMapVariable(featureName, packageName, featurePackage, testPath string) error {
	fset := token.NewFileSet()

	fileAst, err := parser.ParseFile(fset, testPath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	if !astutil.UsesImport(fileAst, featurePackage) {
		astutil.AddImport(fset, fileAst, featurePackage)
	}

	pre := func(c *astutil.Cursor) bool {
		if sel, ok := c.Node().(*ast.GenDecl); ok {
			for _, spec := range sel.Specs {
				spec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for _, fn := range spec.Names {
					if fn.Name != mapVariableName {
						continue
					}

					features := fn.Obj.Decl.(*ast.ValueSpec).Values[0]
					features.(*ast.CompositeLit).Elts = append(features.(*ast.CompositeLit).Elts,
						&ast.KeyValueExpr{
							Key: &ast.BasicLit{
								Kind:  token.STRING,
								Value: fmt.Sprintf("\n\"%v\"", featureName),
							},
							Value: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: packageName,
								},
								Sel: &ast.Ident{
									Name: "InitializeScenario,\n",
								},
							},
						},
					)

					break
				}
			}
		}

		return true
	}

	astutil.Apply(fileAst, pre, nil)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, fileAst); err != nil {
		return err
	}

	return ioutil.WriteFile(testPath, buf.Bytes(), 0644)
}

func generateGoFile(mapping *Mapping) error {
	buf := bytes.NewBuffer(make([]byte, 0))

	err := codeGenTemplate.Execute(buf, mapping)
	if err != nil {
		return err
	}

	// 13. if update is set
	isDirOk := files.IsDir(mapping.GoFile)
	if !isDirOk {
		err := os.MkdirAll(filepath.Dir(mapping.GoFile), 0755)
		if err != nil {
			return err
		}
	}

	return ioutil.WriteFile(mapping.GoFile, buf.Bytes(), 0644)
}

func extractSignatureChanges(mapping *Mapping) []SignatureChange {
	var signatureChanges []SignatureChange
	for _, feature := range mapping.Features {
		for _, gofunc := range mapping.GoDefinitions {
			if feature.Name != gofunc.Name {
				continue
			}

			// We need to compare function arguments checking only
			// the number and type. Is not possible to rely in the name
			// in the go code.
			featKeys := feature.Args.Keys()
			goKeys := gofunc.Args.Keys()

			if len(featKeys) != len(goKeys) {
				signatureChanges = append(signatureChanges, SignatureChange{
					Function: gofunc.Name,
					Have:     argsFromMap(gofunc.Args, true),
					Want:     argsFromMap(feature.Args, true),
				})

				continue
			}

			for index, k := range featKeys {
				fv, _ := feature.Args.Get(k)
				gv, _ := gofunc.Args.Get(goKeys[index])

				if !reflect.DeepEqual(fv, gv) {
					signatureChanges = append(signatureChanges, SignatureChange{
						Function: gofunc.Name,
						Have:     argsFromMap(gofunc.Args, true),
						Want:     argsFromMap(feature.Args, true),
					})
				}
			}
		}
	}

	return signatureChanges
}

// Code below this comment comes from github.com/cucumber/godog
// (code defined in private methods)

const (
	numberGroup = "(\\d+)"
	stringGroup = "\"([^\"]*)\""
)

// parseStepArgs extracts arguments from an expression defined in a step RegExp.
// This code was extracted from
// https://github.com/cucumber/godog/blob/4da503aab2d0b71d380fbe8c48a6af9f729b6f5a/undefined_snippets_gen.go#L41
func parseStepArgs(exp string, argument *messages.PickleStepArgument) *orderedmap.OrderedMap {
	var (
		args      []string
		pos       int
		breakLoop bool
	)

	for !breakLoop {
		part := exp[pos:]
		ipos := strings.Index(part, numberGroup)
		spos := strings.Index(part, stringGroup)

		switch {
		case spos == -1 && ipos == -1:
			breakLoop = true
		case spos == -1:
			pos += ipos + len(numberGroup)
			args = append(args, "int")
		case ipos == -1:
			pos += spos + len(stringGroup)
			args = append(args, "string")
		case ipos < spos:
			pos += ipos + len(numberGroup)
			args = append(args, "int")
		case spos < ipos:
			pos += spos + len(stringGroup)
			args = append(args, "string")
		}
	}

	if argument != nil {
		if argument.GetDocString() != nil {
			args = append(args, "*godog.DocString")
		}

		if argument.GetDataTable() != nil {
			args = append(args, "*godog.DocString")
		}
	}

	stepArgs := orderedmap.New()

	for i, v := range args {
		k := fmt.Sprintf("arg%d", i+1)
		stepArgs.Set(k, v)
	}

	return stepArgs
}

// some snippet formatting regexps
var snippetExprCleanup = regexp.MustCompile("([\\/\\[\\]\\(\\)\\\\^\\$\\.\\|\\?\\*\\+\\'])")
var snippetExprQuoted = regexp.MustCompile("(\\W|^)\"(?:[^\"]*)\"(\\W|$)")
var snippetMethodName = regexp.MustCompile("[^a-zA-Z\\_\\ ]")
var snippetNumbers = regexp.MustCompile("(\\d+)")

// parseSteps converts a string step definition in a different one valid as a regular
// expression that can be used in a go Step definition. This original code is located in
// https://github.com/cucumber/godog/blob/4da503aab2d0b71d380fbe8c48a6af9f729b6f5a/fmt.go#L457
func parseSteps(steps []*messages.Pickle_PickleStep, funcDefs []Function) []Function {
	var index int

	for _, step := range steps {
		text := step.Text

		expr := snippetExprCleanup.ReplaceAllString(text, "\\$1")
		expr = snippetNumbers.ReplaceAllString(expr, "(\\d+)")
		expr = snippetExprQuoted.ReplaceAllString(expr, "$1\"([^\"]*)\"$2")
		expr = "^" + strings.TrimSpace(expr) + "$"

		name := snippetNumbers.ReplaceAllString(text, " ")
		name = snippetExprQuoted.ReplaceAllString(name, " ")
		name = strings.TrimSpace(snippetMethodName.ReplaceAllString(name, ""))

		var words []string
		for i, w := range strings.Split(name, " ") {
			switch {
			case i != 0:
				w = strings.Title(w)
			case len(w) > 0:
				w = string(unicode.ToLower(rune(w[0]))) + w[1:]
			}

			words = append(words, w)
		}

		name = strings.Join(words, "")
		if len(name) == 0 {
			index++
			name = fmt.Sprintf("StepDefinitioninition%d", index)
		}

		var found bool
		for _, f := range funcDefs {
			if f.Expr == expr {
				found = true
				break
			}
		}

		if !found {
			args := parseStepArgs(expr, step.Argument)
			funcDefs = append(funcDefs, Function{Name: name, Expr: expr, Args: args})
		}
	}

	return funcDefs
}
