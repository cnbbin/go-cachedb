package statistic

import (
	"reflect"
	"sync"
	"testing"
)

func TestRegisterAndGetCategories_Static(t *testing.T) {
	handler := StatisticHandler(1)
	statType := StatisticType(100)
	categories := []StatisticTypeCategory{10, 20, 30}

	RegisterCategories(handler, statType, categories)

	got := GetCategories(handler, statType)
	if len(got) != len(categories) {
		t.Fatalf("expected %v categories, got %v", len(categories), len(got))
	}
	for i, c := range got {
		if c != categories[i] {
			t.Errorf("expected category %v at index %d, got %v", categories[i], i, c)
		}
	}
}

func TestRegisterAndGetCategories_QueryFunc(t *testing.T) {
	handler := StatisticHandler(2)
	statType := StatisticType(200)
	queryCalled := false

	RegisterQueryFunc(handler, func(t StatisticType) []StatisticTypeCategory {
		queryCalled = true
		if t == statType {
			return []StatisticTypeCategory{40, 50}
		}
		return nil
	})

	got := GetCategories(handler, statType)
	if !queryCalled {
		t.Error("expected query function to be called")
	}
	if len(got) != 2 || got[0] != 40 || got[1] != 50 {
		t.Errorf("unexpected categories returned: %v", got)
	}

	// 再次调用，测试缓存是否生效，queryCalled 不应增加
	queryCalled = false
	got2 := GetCategories(handler, statType)
	if queryCalled {
		t.Error("query function should NOT be called again due to caching")
	}
	if len(got2) != 2 || got2[0] != 40 || got2[1] != 50 {
		t.Errorf("unexpected categories on second call: %v", got2)
	}
}

func TestRegisterAndGetCategories_WorkerFunc(t *testing.T) {
	handler := StatisticHandler(3)
	statType := StatisticType(300)
	staticCategories := []StatisticTypeCategory{1, 2, 3}

	RegisterCategories(handler, statType, staticCategories)
	RegisterWorkerFunc(handler, func(t StatisticType, cats []StatisticTypeCategory) []StatisticTypeCategory {
		// 反转切片示例
		n := len(cats)
		ret := make([]StatisticTypeCategory, n)
		for i, c := range cats {
			ret[n-1-i] = c
		}
		return ret
	})

	got := GetCategories(handler, statType)
	expected := []StatisticTypeCategory{3, 2, 1}
	if len(got) != len(expected) {
		t.Fatalf("expected %v categories, got %v", len(expected), len(got))
	}
	for i, c := range got {
		if c != expected[i] {
			t.Errorf("expected category %v at index %d, got %v", expected[i], i, c)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	handler := StatisticHandler(4)
	statType := StatisticType(400)
	staticCategories := []StatisticTypeCategory{5, 6, 7}

	RegisterCategories(handler, statType, staticCategories)
	RegisterWorkerFunc(handler, func(t StatisticType, cats []StatisticTypeCategory) []StatisticTypeCategory {
		return cats // 直接返回
	})

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			got := GetCategories(handler, statType)
			if len(got) != len(staticCategories) {
				t.Errorf("expected %v categories, got %v", len(staticCategories), len(got))
			}
		}()
	}

	wg.Wait()
}

const (
	TestHandler StatisticHandler = 1
	TestType    StatisticType    = 100
)

func TestStaticFunc(t *testing.T) {
	// 初始化数据
	initialCats := []StatisticTypeCategory{1, 2, 3}
	RegisterCategories(TestHandler, TestType, initialCats)

	// 添加 staticFunc：将 addValue 加入每个类别中
	RegisterStaticFunc(TestHandler, func(statType StatisticType, categories []StatisticTypeCategory, addValue int32) {
		for i := range categories {
			categories[i] += StatisticTypeCategory(addValue)
		}
	})

	// 执行 staticFunc，增加值 10
	ApplyStaticFunc(TestHandler, TestType, 10)

	// 重新获取静态数据
	got := GetCategories(TestHandler, TestType)

	want := []StatisticTypeCategory{11, 12, 13}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ApplyStaticFunc failed: got %v, want %v", got, want)
	}
}
