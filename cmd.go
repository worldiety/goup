package main

func main() {

	args := &Args{}
	args.Evaluate()

	gp, err := NewGoup(args)
	must(err)

	err = gp.Build()
	must(err)



}
