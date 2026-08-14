package main

type Com struct {
	Base string
	Args map[string][]string
	// for all optional args
	// some commands support optional args
	Extras map[string][]string
}

func GetCom(base string, rest []string) *Com {
	com := &Com{
		Base: base,
	}

	switch base {
	case "ping":
		com.ValidatePing(rest)
	}

	return com
}

func (com *Com) ValidatePing(args []string) {} // eat five star do nothing

func (com *Com) ValidateEcho(args []string) {
}
