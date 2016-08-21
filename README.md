# workman

[godoc](https://godoc.org/github.com/APTy/workman)

Package workman provides constructs for parallel processing.

## Usage

``` go
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
```

## Specifying Workers

Construct a new work manager with the desired number of parallel workers

``` go
wm, err := NewWorkManager(10)
```

## Sending Work

When using the WorkManager, it is important to invoke SendWork as if you were invoking your work function directly. That is, if you start workers using a function with the following signature:

``` go
wm.StartWorkers(func(n int, line string, active bool) {})
```

Then you should send it work as follows:

``` go
wm.SendWork(10, "Townsend", true)
wm.SendWork(10, "Folsom/Pacific", false)
```

## Handling Errors

Your function should return an error or nothing

``` go
wm.StartWorkers(func() error {
    return errors.New("bad run")
})
```

The error returned from WaitForCompletion indicates if at least one error occurred during processing. Those errors can be enumerated as follows:

``` go
if err := wm.WaitForCompletion(); err != nil {
    errs := err.All()
    for _, e := range errs {
        fmt.Println(e)
    }
}
```

Happy processing!
