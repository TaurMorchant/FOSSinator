package processor

import (
	"fmt"
	"fossinator/config"
	"fossinator/fs"
	"go/ast"
	"go/token"
	"os"
)

const PreComment = "//this is autogenerated code with default service loading configuration. Please review it"

func AddConfigLoaderConfiguration(dir string) error {
	fmt.Printf("----- Add Config Loader Configuration [START] -----\n")
	defer fmt.Printf("----- Add Config Loader Configuration [END] -----\n\n")

	mainFileName, err := fs.FindMainFile(dir)
	if err != nil {
		fmt.Println("Main file is mot found => skip step")
		return nil
	}
	fmt.Println("mainFileName = ", mainFileName)

	srcBytes, err := os.ReadFile(mainFileName)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}
	src := string(srcBytes)

	src, err = insertImports(src, config.CurrentConfig.Go.ServiceLoading.Imports)
	if err != nil {
		return err
	}

	src, err = insertInitPart(src, config.CurrentConfig.Go.ServiceLoading.Instructions)
	if err != nil {
		return err
	}

	fmt.Println("Updated:", mainFileName)
	return fs.WriteFile(mainFileName, src)
}

//-------------------------------------------------------------------------

func insertImports(src string, list []string) (string, error) {
	if list == nil || len(list) == 0 {
		return src, nil
	}
	fileSet, file, err := fs.ParseSrc(src)
	if err != nil {
		return "", err
	}

	insertion := formatImports()

	insertPos, importBlock := findInsertImportPosition(fileSet, file, src)

	if importBlock == nil {
		insertion = "import (\n" + insertion + ")\n\n"
	}

	return insertIntoPosition(src, insertion, insertPos), nil
}

func formatImports() string {
	var result string
	for _, imp := range config.CurrentConfig.Go.ServiceLoading.Imports {
		result += "\t\"" + imp + "\"\n"
	}
	return result
}

func findInsertImportPosition(fs *token.FileSet, file *ast.File, src string) (int, *ast.GenDecl) {
	var insertPos int
	var importBlock *ast.GenDecl

	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			importBlock = genDecl
			insertPos = fs.Position(importBlock.Rparen).Offset
			return insertPos, importBlock
		}
	}

	return findFirstFuncPosition(fs, file, src), nil
}

func findFirstFuncPosition(fs *token.FileSet, f *ast.File, src string) int {
	result := len(src)
	for _, decl := range f.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Doc != nil && len(funcDecl.Doc.List) > 0 {
				return fs.Position(funcDecl.Doc.Pos()).Offset
			} else {
				return fs.Position(funcDecl.Pos()).Offset
			}
		}
	}
	return result
}

func insertInitPart(src string, list []string) (string, error) {
	if list == nil || len(list) == 0 {
		return src, nil
	}

	fileSet, file, err := fs.ParseSrc(src)
	if err != nil {
		return "", err
	}

	initFunc := findFunc(file, "init")

	var insertPos int
	insertion := formatInitStatements(list)

	if initFunc == nil {
		insertion = "func init() {\n" + insertion + "}\n\n"

		insertPos = findFirstFuncPosition(fileSet, file, src)
	} else {
		insertPos = fileSet.Position(initFunc.Rbrace).Offset
	}

	return insertIntoPosition(src, insertion, insertPos), nil
}

func formatInitStatements(list []string) string {
	result := "\t" + PreComment + "\n"
	for _, line := range list {
		result += "\t" + line + "\n"
	}
	return result
}

func findFunc(f *ast.File, name string) *ast.BlockStmt {
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == name && fn.Recv == nil {
			return fn.Body
		}
	}
	return nil
}

func insertIntoPosition(src, insertion string, pos int) string {
	return src[:pos] + insertion + src[pos:]
}
