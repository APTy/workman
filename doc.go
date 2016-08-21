/*
Package workman provides constructs for parallel processing.

	getHTTP := func(i int) {
	    res, _ := http.Get(fmt.Sprintf("http://httpbin.org/get?n=%d", i))
	    data, _ := ioutil.ReadAll(res.Body)
	    res.Body.Close()
	    fmt.Printf("%s", data)
	}

	wm, _ := NewWorkManager(4)
	wm.StartWorkers(getHTTP)
	for i := 0; i < 10; i++ {
	    wm.SendWork(i)
	}
	wm.WaitForCompletion()

Construct a new work manager with the desired number of parallel workers

	wm, err := NewWorkManager(10)

When using the WorkManager, it is important to invoke SendWork as if you were
invoking your work function directly. That is, if you start workers using a
function with the following signature:

	wm.StartWorkers(func(n int, line string, active bool) {})

Then you should send it work as follows:

	wm.SendWork(10, "Townsend", true)
	wm.SendWork(10, "Folsom/Pacific", false)

Your function should return an error or nothing

	wm.StartWorkers(func() error {
		return errors.New("bad run")
	})

The error returned from WaitForCompletion indicates if at least one error occurred
during processing. Those errors can be enumerated as follows:

	if err := wm.WaitForCompletion(); err != nil {
		errs := err.All()
		for _, e := range errs {
			fmt.Println(e)
		}
	}

Happy processing!
*/
package workman
