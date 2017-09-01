package workman

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoArgs(t *testing.T) {
	work := func() error {
		return nil
	}

	wm, err := NewWorkManager(10)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	for i := 0; i < 9999; i++ {
		err := wm.SendWork()
		assert.Nil(t, err)
	}
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
}

func TestSingleArg(t *testing.T) {
	list := []int{1, 2, 3, 4, 5}
	work := func(i int) error {
		return nil
	}

	wm, err := NewWorkManager(10)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	for _, i := range list {
		err := wm.SendWork(i)
		assert.Nil(t, err)
	}
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
}

func TestMultipleArgs(t *testing.T) {
	list1 := []int{1, 2, 3, 4, 5}
	list2 := []string{"a", "b", "c", "d", "e"}
	work := func(i int, s string) error {
		return nil
	}

	wm, err := NewWorkManager(10)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	for i := range list1 {
		err := wm.SendWork(list1[i], list2[i])
		assert.Nil(t, err)
	}
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
}

func TestStructLiteralArg(t *testing.T) {
	type T struct {
		s string
		n int
	}
	list := []T{{"a", 1}, {"b", 2}, {"c", 3}, {"d", 4}, {"e", 5}}
	work := func(t T) error {
		return nil
	}

	wm, err := NewWorkManager(1)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	for i := range list {
		err := wm.SendWork(list[i])
		assert.Nil(t, err)
	}
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
}

func TestStructPtrArg(t *testing.T) {
	type T struct {
		s string
		n int
	}
	list := []T{{"a", 1}, {"b", 2}, {"c", 3}, {"d", 4}, {"e", 5}}
	work := func(t *T) error {
		return nil
	}

	wm, err := NewWorkManager(5)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	for i := range list {
		err := wm.SendWork(&list[i])
		assert.Nil(t, err)
	}
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
}

func TestMutateStructPtrArg(t *testing.T) {
	type T struct {
		s string
		n int
	}
	list := []T{{"a", 1}, {"b", 2}, {"c", 3}, {"d", 4}, {"e", 5}}
	work := func(t *T) error {
		t.s = "test"
		t.n = 1337
		return nil
	}

	wm, err := NewWorkManager(5)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	for i := range list {
		err := wm.SendWork(&list[i])
		assert.Nil(t, err)
	}
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
	for i := range list {
		assert.Equal(t, list[i].s, "test")
		assert.Equal(t, list[i].n, 1337)
	}
}

func TestPoolSize(t *testing.T) {
	for i := 1; i < 1000; i++ {
		list1 := []int{1, 2, 3, 4, 5}
		list2 := []string{"a", "b", "c", "d", "e"}
		work := func(i int, s string) error {
			return nil
		}

		wm, err := NewWorkManager(2)
		assert.Nil(t, err)
		wm.StartWorkers(work)
		for i := range list1 {
			err := wm.SendWork(list1[i], list2[i])
			assert.Nil(t, err)
		}
		err = wm.WaitForCompletion()
		assert.Nil(t, err)
	}
}

func TestReturnedErrors(t *testing.T) {
	errEven := errors.New("arg is even")
	list := []int{1, 2, 3, 4, 5}
	work := func(i int) error {
		if i%2 == 0 {
			return errEven
		}
		return nil
	}

	wm, err := NewWorkManager(100)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	for i := range list {
		err := wm.SendWork(list[i])
		assert.Nil(t, err)
	}
	errs := wm.WaitForCompletion()
	assert.EqualError(t, errs, ErrWorkerErrors.Error())

	wErrs := errs.All()
	assert.Equal(t, len(wErrs), 2, "It should return 2 errors")
	assert.EqualError(t, wErrs[0], errEven.Error())
	assert.EqualError(t, wErrs[1], errEven.Error())
}

func TestBadArgs(t *testing.T) {
	var wm WorkManager
	var err error
	work := func(i int) error {
		return nil
	}

	wm, err = NewWorkManager(0)
	assert.EqualError(t, err, ErrTooFewWorkers.Error())

	wm, err = NewWorkManager(1)
	err = wm.StartWorkers(1)
	assert.EqualError(t, err, ErrInvalidWorkFuncType.Error())

	err = wm.StartWorkers(nil)
	assert.EqualError(t, err, ErrInvalidWorkFuncType.Error())

	err = wm.StartWorkers(work)
	assert.Nil(t, err)

	err = wm.WaitForCompletion()
	assert.Nil(t, err)

	err = wm.StartWorkers(work)
	assert.EqualError(t, err, ErrAlreadyStarted.Error())

	err = wm.WaitForCompletion()
	assert.EqualError(t, err, ErrWorkCompleted.Error())

	err = wm.SendWork(1)
	assert.EqualError(t, err, ErrWorkCompleted.Error())
}

func TestNoWorkReturn(t *testing.T) {
	list := []int{1, 2, 3, 4, 5}
	work := func(i int) {}

	wm, err := NewWorkManager(2)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	for _, i := range list {
		err := wm.SendWork(i)
		assert.Nil(t, err)
	}
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
}

func TestNonErrorWorkReturn(t *testing.T) {
	type nonErr interface{}
	work := func() nonErr { return "test" }

	wm, err := NewWorkManager(1)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	err = wm.SendWork()
	assert.Nil(t, err)
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
}

func TestNumArgsMismatch(t *testing.T) {
	list := []int{1, 2, 3, 4, 5}
	work := func(i int) {}

	wm, err := NewWorkManager(2)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	for _, i := range list {
		err := wm.SendWork(i, "badarg")
		assert.EqualError(t, err, ErrBadWorkArgs.Error())
	}
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
}

func TestArgTypeMismatch(t *testing.T) {
	work := func(i int) {}

	wm, err := NewWorkManager(2)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	err = wm.SendWork("badarg")
	assert.EqualError(t, err, ErrBadWorkArgs.Error())
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
}

type myInterface interface {
	DoWork()
}
type myStruct struct {}
func (s *myStruct) DoWork() {}
func TestInterfaceArg(t *testing.T) {
	work := func(s myInterface) {}

	ms := &myStruct{}
	wm, err := NewWorkManager(2)
	assert.Nil(t, err)
	wm.StartWorkers(work)
	err = wm.SendWork(ms)
	assert.Nil(t, err)
	err = wm.WaitForCompletion()
	assert.Nil(t, err)
}

func TestWorkerTimeout(t *testing.T) {
	work := func() { select {} }

	wm, err := NewWorkManager(20)
	wm.WorkerMaxTimeout = 10 * time.Millisecond

	assert.Nil(t, err)
	wm.StartWorkers(work)
	for i := 0; i < 100; i++ {
		err = wm.SendWork()
		assert.Nil(t, err)
	}
	errs := wm.WaitForCompletion()
	assert.EqualError(t, errs, ErrWorkerErrors.Error())
	assert.EqualError(t, errs.All()[0], ErrWorkerTimeout.Error())

}
