package main

func main() {

	args := &Args{}
	args.Evaluate()

	_, err := NewGoup(args)
	must(err)


	/*

		total := time.Now()
		start := time.Now()

		builder := &Builder{}
		builder.Init()

		if !builder.IsBuildRequired() {
			builder.PP("everything is up to date, nothing to do")
			builder.StopWatch(total, "total")
			os.Exit(0)
		}


		builder.EnsureGoMobile()
		builder.StopWatch(start, "preparation")

		start = time.Now()
		err := builder.CopyModulesToWorkspace()
		if err != nil {
			fmt.Println("failed to prepare modules in workspace:", err)
			os.Exit(-1)
		}
		builder.StopWatch(start, "workspace setup")

		start = time.Now()
		err = builder.Gomobile()
		if err != nil {
			fmt.Println("failed to compile with gomobile:", err)
			os.Exit(-1)
		}
		builder.UpdateBuildCache()

		builder.StopWatch(start, "gomobile")
		builder.StopWatch(total, "total")

	*/

}
