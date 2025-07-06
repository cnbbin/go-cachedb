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

func TestConcurrentRegisterAndQuery(t *testing.T) {
	handler := StatisticHandler(7)
	statType := StatisticType(700)

	var wg sync.WaitGroup
	const goroutines = 100

	// 并发注册不同的 categories
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cat := StatisticTypeCategory(i)
			RegisterCategories(handler, statType, []StatisticTypeCategory{cat})
		}(i)
	}

	// 同时并发查询
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = GetCategories(handler, statType)
		}()
	}

	wg.Wait()

	// 最终检查所有 category 是否都被注册（允许重复）
	got := GetCategories(handler, statType)
	seen := make(map[StatisticTypeCategory]bool)
	for _, c := range got {
		seen[c] = true
	}

	if len(seen) != goroutines {
		t.Errorf("expected %d unique categories, got %d", goroutines, len(seen))
	}
}

func TestInitialStaticAndQueryFuncPriority(t *testing.T) {
	handler := StatisticHandler(8)
	statType := StatisticType(800)

	// 1. 注册静态初始值
	initialCats := []StatisticTypeCategory{1000, 2000, 3000}
	RegisterCategories(handler, statType, initialCats)

	// 2. 注册动态查询函数（理论上不会调用）
	queryCalled := false
	RegisterQueryFunc(handler, func(t StatisticType) []StatisticTypeCategory {
		queryCalled = true
		return []StatisticTypeCategory{4000, 5000}
	})

	// 3. 查询时应直接返回静态值，动态查询函数不应该被调用
	cats := GetCategories(handler, statType)
	if queryCalled {
		t.Error("query function should NOT be called when static cache exists")
	}

	if len(cats) != len(initialCats) {
		t.Fatalf("expected %d categories, got %d", len(initialCats), len(cats))
	}
	for i, c := range cats {
		if c != initialCats[i] {
			t.Errorf("expected category %v at index %d, got %v", initialCats[i], i, c)
		}
	}

	// 4. 再测试一个没有静态缓存的类型，动态查询函数应该被调用
	missingType := StatisticType(801)
	queryCalled = false
	cats2 := GetCategories(handler, missingType)
	if !queryCalled {
		t.Error("query function should be called when static cache missing")
	}
	expectedQueryCats := []StatisticTypeCategory{4000, 5000}
	if len(cats2) != len(expectedQueryCats) {
		t.Fatalf("expected %d categories from query, got %d", len(expectedQueryCats), len(cats2))
	}
	for i, c := range cats2 {
		if c != expectedQueryCats[i] {
			t.Errorf("expected query category %v at index %d, got %v", expectedQueryCats[i], i, c)
		}
	}
}

func TestDynamicQueryFuncRegistration(t *testing.T) {
	handler := StatisticHandler(9)
	statType := StatisticType(900)

	var callCount int
	var mu sync.Mutex

	// 动态注册第一个查询函数
	RegisterQueryFunc(handler, func(t StatisticType) []StatisticTypeCategory {
		mu.Lock()
		callCount++
		mu.Unlock()
		if t == statType {
			return []StatisticTypeCategory{1, 2, 3}
		}
		return nil
	})

	// 查询，应该调用动态查询函数
	cats := GetCategories(handler, statType)
	if len(cats) != 3 {
		t.Errorf("expected 3 categories, got %v", cats)
	}

	// 再次查询，应该使用缓存，不会增加调用次数
	mu.Lock()
	callCountBefore := callCount
	mu.Unlock()

	// 再次查询，应该使用缓存，不会调用动态查询函数
	cats2 := GetCategories(handler, statType)

	if len(cats2) != 3 {
		t.Errorf("expected 3 categories on second call, got %v", cats2)
	}

	for i, c := range cats2 {
		if c != []StatisticTypeCategory{1, 2, 3}[i] {
			t.Errorf("expected category %v at index %d, got %v", []StatisticTypeCategory{1, 2, 3}[i], i, c)
		}
	}
	mu.Lock()
	callCountAfter := callCount
	mu.Unlock()

	if callCountAfter != callCountBefore {
		t.Errorf("expected no new call to query func, but got increase")
	}

	// 清理静态缓存（直接删除对应的 map 项）
	mgr := GetGlobalManager()
	mgr.mu.Lock()
	if hInfo, exists := mgr.registries[handler]; exists {
		delete(hInfo.staticInfo, statType)
	}
	mgr.mu.Unlock()

	// 重新注册查询函数，覆盖前一个
	RegisterQueryFunc(handler, func(t StatisticType) []StatisticTypeCategory {
		return []StatisticTypeCategory{4, 5, 6}
	})

	// 再次查询，应该调用新的动态查询函数，返回新的结果
	cats3 := GetCategories(handler, statType)
	expected := []StatisticTypeCategory{4, 5, 6}
	if len(cats3) != len(expected) {
		t.Errorf("expected new categories %v, got %v", expected, cats3)
	}
	for i := range expected {
		if cats3[i] != expected[i] {
			t.Errorf("expected %v at index %d, got %v", expected[i], i, cats3[i])
		}
	}
}

func TestDynamicQueryFuncConcurrent(t *testing.T) {
	handler := StatisticHandler(10)
	statType := StatisticType(1000)

	// 注册动态查询函数
	RegisterQueryFunc(handler, func(t StatisticType) []StatisticTypeCategory {
		return []StatisticTypeCategory{100, 200}
	})

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	// 并发调用 GetCategories
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			cats := GetCategories(handler, statType)
			if len(cats) != 2 {
				t.Errorf("expected 2 categories, got %v", cats)
			}
		}()
	}

	wg.Wait()
}
