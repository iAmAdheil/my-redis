package main

import "errors"

type Com struct {
	Base string
	Args map[string][]string
	// for all optional args
	// some commands support optional args
	Extras map[string][]string
}

func GetCom(parts []string) (*Com, error) {
	base := parts[1]
	rest := parts[2:]

	com := &Com{
		Base: base,
	}

	var err error

	switch base {
	case "ping":
		com.ValidatePing(rest)
	case "echo":
		err = com.ValidateEcho(rest)
	}

	return com, err
}

func (com *Com) ValidatePing(args []string) {} // eat five star do nothing

func (com *Com) ValidateEcho(args []string) error {
	if len(args) >= 2 {
		com.Args["echothis"] = []string{args[1]}
	} else {
		return errors.New("Echo expects atleast one argument")
	}

	return nil
}
