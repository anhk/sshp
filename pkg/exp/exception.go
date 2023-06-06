package exp

func Throw(e any) {
	if e != nil {
		panic(e)
	}
}

func TryCatch(try func(), catch func(e any)) {
	defer func() {
		if e := recover(); e != nil {
			catch(e)
		}
	}()
	try()
}

func Try(try func()) any {
	var ret any = nil
	TryCatch(try, func(e any) {
		ret = e
	})
	return ret
}
