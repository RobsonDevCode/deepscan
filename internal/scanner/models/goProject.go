package scannermodels

import "golang.org/x/mod/modfile"

type GoProject struct {
	ServiceName string
	LangVersion string
	Packages    []*modfile.Require
}
