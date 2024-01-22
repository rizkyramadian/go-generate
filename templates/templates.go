package templates

func InitFile(n string, r string, pkgs string) string {
	res :=
		`package ` + n + `  

type ` + n + `Impl struct {
	//TODO: Insert Dependencies Here
}

//TODO: Insert New Func here
func New() ` + pkgs + "." + r + `{
	return &` + n + `Impl{}
}

`
	return res
}
