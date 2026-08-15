package main

import "errors"

type Com struct {
	Base string
	Args map[string][]string
	// for all optional args
	// some commands support optional args
	Extras map[string][]string
}

func GetCom(base string, args []string) (*Com, error) {
	com := &Com{
		Base: base,
	}

	var err error

	switch base {
	case "ping":
		com.ValidatePing(args)
	case "echo":
		err = com.ValidateEcho(args)
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
