package examples

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

func testHTTP() {
	wm, _ := NewWorkManager(4)
	wm.StartWorkers(func(i int) error {
		res, err := http.Get(fmt.Sprintf("http://httpbin.org/get?n=%d", i))
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		res.Body.Close()
		fmt.Printf("%s", data)
		return nil
	})

	for _, i := range []int{1, 2, 3, 4, 5} {
		wm.SendWork(i)
	}

	wm.WaitForCompletion()

}

func testErr() {
	wm, _ := NewWorkManager(2)
	wm.StartWorkers(func() error {
		return errors.New("bad run")
	})
	for i := 0; i < 4; i++ {
		wm.SendWork()
	}
	if err := wm.WaitForCompletion(); err != nil {
		errs := err.All()
		for _, e := range errs {
			fmt.Println(e)
		}
	}
}
