package keylock

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRead_SingleExecutionPerKey(t *testing.T) {
	kl := &KeyLock[int]{locks: make(map[string]*status[int])}

	var calls int32

	fn := func() (int, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond)
		return 42, nil
	}

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	results := make([]int, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			v, err := kl.Read("key", fn)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			results[i] = v
		}(i)
	}

	wg.Wait()

	if calls != 1 {
		t.Fatalf("fn executed %d times, want 1", calls)
	}

	for i, v := range results {
		if v != 42 {
			t.Fatalf("result[%d]=%d, want 42", i, v)
		}
	}
}

func TestRead_WaitersBlockUntilDone(t *testing.T) {
	kl := &KeyLock[int]{locks: make(map[string]*status[int])}

	started := make(chan struct{})
	release := make(chan struct{})

	fn := func() (int, error) {
		close(started)
		<-release
		return 7, nil
	}

	go func() {
		_, _ = kl.Read("key", fn)
	}()

	<-started

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = kl.Read("key", fn)
	}()

	select {
	case <-done:
		t.Fatal("second Read returned before first completed")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("second Read did not return after release")
	}
}

func TestWrite_MutualExclusion(t *testing.T) {
	kl := &KeyLock[int]{locks: make(map[string]*status[int])}

	var active int32
	var maxActive int32

	fn := func() {
		n := atomic.AddInt32(&active, 1)
		if n > maxActive {
			atomic.StoreInt32(&maxActive, n)
		}
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&active, -1)
	}

	const goroutines = 5
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			kl.Write("key", fn)
		}()
	}

	wg.Wait()

	if maxActive != 1 {
		t.Fatalf("max concurrent writers = %d, want 1", maxActive)
	}
}

func TestReadAndWrite_Exclusive(t *testing.T) {
	kl := &KeyLock[int]{locks: make(map[string]*status[int])}

	var order []string
	var mu sync.Mutex

	readFn := func() (int, error) {
		mu.Lock()
		order = append(order, "read")
		mu.Unlock()
		time.Sleep(30 * time.Millisecond)
		return 1, nil
	}

	writeFn := func() {
		mu.Lock()
		order = append(order, "write")
		mu.Unlock()
		time.Sleep(30 * time.Millisecond)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = kl.Read("key:r", readFn)
	}()

	go func() {
		defer wg.Done()
		kl.Write("key:w", writeFn)
	}()

	wg.Wait()

	if len(order) != 2 {
		fmt.Println(order)
		t.Fatalf("unexpected execution order: %v", order)
	}
}

func TestPanicInRead_DoesNotDeadlock(t *testing.T) {
	kl := &KeyLock[int]{locks: make(map[string]*status[int])}

	panicFn := func() (int, error) {
		panic("boom")
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		_, _ = kl.Read("key", panicFn)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("panic caused deadlock")
	}
}

func BenchmarkRead_SingleKey(b *testing.B) {
	kl := &KeyLock[int]{locks: make(map[string]*status[int])}

	fn := func() (int, error) {
		return 1, nil
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = kl.Read("key", fn)
		}
	})
}

func BenchmarkRead_MultiKey(b *testing.B) {
	kl := &KeyLock[int]{locks: make(map[string]*status[int])}

	fn := func() (int, error) {
		return 1, nil
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := "key-" + string(rune('a'+(i%26)))
			_, _ = kl.Read(key, fn)
			i++
		}
	})
}

func BenchmarkWrite_SingleKey(b *testing.B) {
	kl := &KeyLock[int]{locks: make(map[string]*status[int])}

	fn := func() {}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			kl.Write("key", fn)
		}
	})
}
