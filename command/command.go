package command

type Command struct {
	Exec  func(*Context, []string)
	Usage string
}

var Commands = map[string]Command{
	"modping": Command{
		Exec:  modping,
		Usage: modpingUsage,
	},
}
