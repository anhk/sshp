package exp

func Throw(e any) {
	if e != nil {
		panic(e)
	}
}
