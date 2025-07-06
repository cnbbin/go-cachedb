package statistic

import (
	"reflect"
	"sync"
	"testing"
)

func TestStaticDoubleFunc(t *testing.T) {
	handler := StatisticHandler(5)
	statType := StatisticType(500)
	initialCats := []StatisticTypeCategory{10, 20}

	// 注册静态分类
	RegisterCategories(handler, statType, initialCats)

	// 注册 workerDoubleFunc，加上两个值并记录被调用次数
	workerCalled := false
	RegisterWorkerDoubleFunc(handler, func(t StatisticType, cats []StatisticTypeCategory, v1, v2 int32) {
		workerCalled = true
		for i := range cats {
			cats[i] += StatisticTypeCategory(v1 + v2)
		}
	})

	// 注册 staticDoubleFunc，验证是否调用并进一步修改
	staticCalled := false
	RegisterStaticDoubleFunc(handler, func(t StatisticType, cats []StatisticTypeCategory, v1, v2 int32) {
		staticCalled = true
		for i := range cats {
			cats[i] += 1 // 额外 +1，便于测试
		}
	})

	// 执行 Apply
	ApplyStaticDoubleFunc(handler, statType, 3, 4)

	// 校验是否都被调用
	if !workerCalled {
		t.Error("workerDoubleFunc was not called")
	}
	if !staticCalled {
		t.Error("staticDoubleFunc was not called")
	}

	// 最终应为：原始 [10,20] → workerDoubleFunc 加 7 → [17,27] → staticDoubleFunc 加 1 → [18,28]
	got := GetCategories(handler, statType)
	want := []StatisticTypeCategory{18, 28}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected result: got %v, want %v", got, want)
	}
}

func TestStaticDoubleFunc_Concurrent(t *testing.T) {
	handler := StatisticHandler(6)
	statType := StatisticType(600)
	initialCats := []StatisticTypeCategory{100, 200, 300}

	// 注册静态分类
	RegisterCategories(handler, statType, initialCats)

	// 使用 Mutex 保护共享写标记
	var mu sync.Mutex
	callLog := make([]string, 0)

	RegisterWorkerDoubleFunc(handler, func(t StatisticType, cats []StatisticTypeCategory, v1, v2 int32) {
		mu.Lock()
		callLog = append(callLog, "worker")
		mu.Unlock()

		for i := range cats {
			cats[i] += StatisticTypeCategory(v1 + v2)
		}
	})

	RegisterStaticDoubleFunc(handler, func(t StatisticType, cats []StatisticTypeCategory, v1, v2 int32) {
		mu.Lock()
		callLog = append(callLog, "static")
		mu.Unlock()

		for i := range cats {
			cats[i] += 1 // 稍作修改验证结果
		}
	})

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			ApplyStaticDoubleFunc(handler, statType, 1, 2) // 每次 +3 再 +1
		}()
	}
	wg.Wait()

	// 只验证：不会 panic，且 GetCategories 最终值符合 100 goroutines * 4 增量
	expected := []StatisticTypeCategory{
		initialCats[0] + 4*goroutines,
		initialCats[1] + 4*goroutines,
		initialCats[2] + 4*goroutines,
	}
	got := GetCategories(handler, statType)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("concurrent ApplyStaticDoubleFunc failed: got %v, want %v", got, expected)
	}

	// 验证调用日志数量
	mu.Lock()
	defer mu.Unlock()
	workerCount := 0
	staticCount := 0
	for _, entry := range callLog {
		if entry == "worker" {
			workerCount++
		} else if entry == "static" {
			staticCount++
		}
	}
	if workerCount != goroutines || staticCount != goroutines {
		t.Errorf("expected %d calls to both funcs, got worker=%d static=%d", goroutines, workerCount, staticCount)
	}
}
