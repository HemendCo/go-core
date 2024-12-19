package cli_models

type Info struct {
	Name  string
	Short string
	Long  string
}

type Command struct {
	Name   string
	Short  string
	Long   string
	Action func()
}
