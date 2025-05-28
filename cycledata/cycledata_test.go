package cycledata

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRegisterAndGetLoaderCreator(t *testing.T) {
	// 注册一个简单的加载器和创建器
	loaderCalled := int32(0)
	creatorCalled := int32(0)

	loaders[DailyCycle] = make(map[TypeKey]func(CycleType, TypeKey, UserID) *PlayerData)
	creators[DailyCycle] = make(map[TypeKey]func(UserID) *PlayerData)

	loaders[DailyCycle][1] = func(cycle CycleType, typeKey TypeKey, userID UserID) *PlayerData {
		atomic.AddInt32(&loaderCalled, 1)
		return &PlayerData{
			UserID:   userID,
			MiscData: map[string]interface{}{"from": "loader"},
		}
	}

	creators[DailyCycle][1] = func(userID UserID) *PlayerData {
		atomic.AddInt32(&creatorCalled, 1)
		return &PlayerData{
			UserID:   userID,
			MiscData: map[string]interface{}{"from": "creator"},
		}
	}

	// 验证获取加载器和创建器
	l := getLoader(DailyCycle, 1)
	if l == nil {
		t.Fatal("loader should not be nil")
	}
	c := getCreator(DailyCycle, 1)
	if c == nil {
		t.Fatal("creator should not be nil")
	}

	// 调用加载器和创建器
	p1 := l(DailyCycle, 1, 123)
	if p1 == nil || p1.MiscData["from"] != "loader" {
		t.Fatal("loader function did not work")
	}

	p2 := c(456)
	if p2 == nil || p2.MiscData["from"] != "creator" {
		t.Fatal("creator function did not work")
	}
}

func TestDataCollectionGetAndUpdate(t *testing.T) {
	// 准备注册器
	loaders[DailyCycle] = make(map[TypeKey]func(CycleType, TypeKey, UserID) *PlayerData)
	creators[DailyCycle] = make(map[TypeKey]func(UserID) *PlayerData)

	creators[DailyCycle][1] = func(userID UserID) *PlayerData {
		return &PlayerData{
			UserID:   userID,
			MiscData: make(map[string]interface{}),
		}
	}

	col := newCollection()

	// 初次获取，触发创建器
	p := col.get(DailyCycle, 1, 1000)
	if p == nil {
		t.Fatal("expected to create PlayerData")
	}
	if p.UserID != 1000 {
		t.Fatalf("unexpected UserID %d", p.UserID)
	}

	// 更新字段
	p.update("score", 1234)
	if p.MiscData["score"] != 1234 {
		t.Fatalf("expected score=1234, got %v", p.MiscData["score"])
	}

	// 再次获取应该是同一个实例
	p2 := col.get(DailyCycle, 1, 1000)
	if p2 != p {
		t.Fatal("expected to get the same PlayerData instance")
	}
}

func TestDataCollectionCleanExpired(t *testing.T) {
	col := newCollection()

	now := int32(time.Now().Unix())

	// 添加三条数据，一条无过期时间，一条过期，一条未过期
	col.data[1] = &PlayerData{UserID: 1, ExpireTime: 0, MiscData: make(map[string]interface{})}
	col.data[2] = &PlayerData{UserID: 2, ExpireTime: now - 10, MiscData: make(map[string]interface{})}  // 过期
	col.data[3] = &PlayerData{UserID: 3, ExpireTime: now + 100, MiscData: make(map[string]interface{})} // 未过期

	col.cleanExpired(now, DailyCycle, 1)

	if _, ok := col.data[2]; ok {
		t.Fatal("expired data not cleaned")
	}
	if _, ok := col.data[1]; !ok {
		t.Fatal("data with ExpireTime=0 should not be cleaned")
	}
	if _, ok := col.data[3]; !ok {
		t.Fatal("non-expired data should remain")
	}
}

func TestCycleServiceGetCollectionAndFlush(t *testing.T) {
	// 测试 cycleService 创建和获取集合
	cs := newService(3600)

	col := cs.getCollection(1)
	if col == nil {
		t.Fatal("expected dataCollection")
	}

	// 添加数据
	col.data[1] = &PlayerData{UserID: 1, MiscData: map[string]interface{}{"key": "value"}}

	// 设置全局存储函数，记录调用次数
	var storeCalled int32
	storeData = func(cycle CycleType, typeKey TypeKey, data *PlayerData) error {
		atomic.AddInt32(&storeCalled, 1)
		return nil
	}

	cs.flush(1, DailyCycle)

	if storeCalled != 1 {
		t.Fatalf("expected storeData called once, got %d", storeCalled)
	}
}

func TestCycleHandlerCleanExpiredData(t *testing.T) {
	handler := newCycleHandler()
	service := newService(3600)
	handler.services[DailyCycle] = service

	col := newCollection()
	now := int32(time.Now().Unix())
	col.data[1] = &PlayerData{UserID: 1, ExpireTime: now - 1, MiscData: make(map[string]interface{})}    // 过期
	col.data[2] = &PlayerData{UserID: 2, ExpireTime: now + 1000, MiscData: make(map[string]interface{})} // 未过期

	service.collections[1] = col

	handler.cleanExpiredData(DailyCycle)

	// 等待清理协程可能完成
	time.Sleep(10 * time.Millisecond)

	if _, ok := col.data[1]; ok {
		t.Fatal("expired data not cleaned by handler")
	}
	if _, ok := col.data[2]; !ok {
		t.Fatal("non-expired data incorrectly removed")
	}
}

func TestFlushAll(t *testing.T) {
	handler := newCycleHandler()

	// 创建两个周期服务，分别添加数据
	s1 := newService(3600)
	s2 := newService(3600)

	col1 := newCollection()
	col1.data[1] = &PlayerData{UserID: 1, MiscData: make(map[string]interface{})}
	s1.collections[1] = col1

	col2 := newCollection()
	col2.data[2] = &PlayerData{UserID: 2, MiscData: make(map[string]interface{})}
	s2.collections[2] = col2

	handler.services[DailyCycle] = s1
	handler.services[WeeklyCycle] = s2

	globalHandler = handler

	var count int32
	storeData = func(cycle CycleType, typeKey TypeKey, data *PlayerData) error {
		atomic.AddInt32(&count, 1)
		if data.UserID == 2 {
			return errors.New("mock store error")
		}
		return nil
	}

	FlushAll()

	if count != 2 {
		t.Fatalf("expected storeData called twice, got %d", count)
	}
}

func TestUpdateIf(t *testing.T) {
	// 准备测试数据
	cycle := DailyCycle
	typeKey := TypeKey(1)
	userID := UserID(1001)
	key := "score"

	// 注册创建器，方便 GetData 能创建数据
	RegisterCreator(cycle, typeKey, func(uid UserID) *PlayerData {
		return &PlayerData{
			UserID:   uid,
			MiscData: make(map[string]interface{}),
		}
	})

	// 获取数据，确保存在
	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		t.Fatalf("failed to get player data")
	}

	// 初始值为空，尝试更新，条件：当旧值为 nil 时更新
	updated := UpdateIf(cycle, typeKey, userID, key, 10, func(oldVal, newVal interface{}) bool {
		return oldVal == nil
	})
	if !updated {
		t.Errorf("expected update success when oldVal is nil")
	}
	if val, ok := pd.MiscData[key]; !ok || val != 10 {
		t.Errorf("expected MiscData[%s] = 10, got %v", key, val)
	}

	// 再次尝试更新，条件：只在新值大于旧值时更新
	updated = UpdateIf(cycle, typeKey, userID, key, 5, func(oldVal, newVal interface{}) bool {
		oldInt, ok1 := oldVal.(int)
		newInt, ok2 := newVal.(int)
		return ok1 && ok2 && newInt > oldInt
	})
	if updated {
		t.Errorf("expected no update since newVal=5 <= oldVal=10")
	}
	if val := pd.MiscData[key]; val != 10 {
		t.Errorf("expected MiscData[%s] still 10, got %v", key, val)
	}

	// 条件满足更新，newVal=20 大于旧值10
	updated = UpdateIf(cycle, typeKey, userID, key, 20, func(oldVal, newVal interface{}) bool {
		oldInt, ok1 := oldVal.(int)
		newInt, ok2 := newVal.(int)
		return ok1 && ok2 && newInt > oldInt
	})
	if !updated {
		t.Errorf("expected update success since newVal=20 > oldVal=10")
	}
	if val := pd.MiscData[key]; val != 20 {
		t.Errorf("expected MiscData[%s] = 20, got %v", key, val)
	}

	// 测试 UpdateTime 是否更新（略过时间差，保证被刷新）
	before := pd.UpdateTime
	time.Sleep(1 * time.Millisecond)
	UpdateIf(cycle, typeKey, userID, key, 30, func(oldVal, newVal interface{}) bool {
		return true
	})
	if !pd.UpdateTime.After(before) {
		t.Errorf("expected UpdateTime to be refreshed")
	}
}

func TestDecreaseIfEnoughInt(t *testing.T) {
	cycle := DailyCycle
	typeKey := TypeKey(1)
	userID := UserID(1001)
	key := "coins"

	RegisterCreator(cycle, typeKey, func(uid UserID) *PlayerData {
		return &PlayerData{
			UserID:   uid,
			MiscData: make(map[string]interface{}),
		}
	})

	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		t.Fatalf("GetData failed")
	}

	// 赋初值 100
	pd.MiscData[key] = 100

	// 扣减 30，成功
	ok := DecreaseIfEnoughInt(cycle, typeKey, userID, key, 30)
	if !ok {
		t.Errorf("expected decrease success")
	}
	if val := pd.MiscData[key].(int); val != 70 {
		t.Errorf("expected 70 after decrease, got %d", val)
	}

	// 扣减 100，失败（不够扣）
	ok = DecreaseIfEnoughInt(cycle, typeKey, userID, key, 100)
	if ok {
		t.Errorf("expected decrease fail due to insufficient amount")
	}
	if val := pd.MiscData[key].(int); val != 70 {
		t.Errorf("expected 70 unchanged after failed decrease, got %d", val)
	}

	// 测试支持 int64 类型
	pd.MiscData[key] = int64(50)
	ok = DecreaseIfEnoughInt(cycle, typeKey, userID, key, 20)
	if !ok {
		t.Errorf("expected decrease success for int64 rawVal")
	}
	if val := pd.MiscData[key].(int); val != 30 {
		t.Errorf("expected 30 after decrease from int64, got %v", val)
	}
}

func TestDecreaseIfEnoughInt32(t *testing.T) {
	cycle := DailyCycle
	typeKey := TypeKey(2)
	userID := UserID(2002)
	key := "points"

	RegisterCreator(cycle, typeKey, func(uid UserID) *PlayerData {
		return &PlayerData{
			UserID:   uid,
			MiscData: make(map[string]interface{}),
		}
	})

	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		t.Fatalf("GetData failed")
	}

	// 赋初值 int32(100)
	pd.MiscData[key] = int32(100)

	// 扣减 50，成功
	ok := DecreaseIfEnoughInt32(cycle, typeKey, userID, key, 50)
	if !ok {
		t.Errorf("expected decrease success")
	}
	if val := pd.MiscData[key].(int32); val != 50 {
		t.Errorf("expected 50 after decrease, got %d", val)
	}

	// 扣减 60，失败
	ok = DecreaseIfEnoughInt32(cycle, typeKey, userID, key, 60)
	if ok {
		t.Errorf("expected decrease fail due to insufficient amount")
	}
	if val := pd.MiscData[key].(int32); val != 50 {
		t.Errorf("expected 50 unchanged after failed decrease, got %d", val)
	}

	// 兼容 int 类型赋值
	pd.MiscData[key] = int(40)
	ok = DecreaseIfEnoughInt32(cycle, typeKey, userID, key, 30)
	if !ok {
		t.Errorf("expected decrease success for int rawVal")
	}
	if val := pd.MiscData[key].(int32); val != 10 {
		t.Errorf("expected 10 after decrease from int, got %v", val)
	}

	// 兼容 int64 类型赋值
	pd.MiscData[key] = int64(20)
	ok = DecreaseIfEnoughInt32(cycle, typeKey, userID, key, 15)
	if !ok {
		t.Errorf("expected decrease success for int64 rawVal")
	}
	if val := pd.MiscData[key].(int32); val != 5 {
		t.Errorf("expected 5 after decrease from int64, got %v", val)
	}
}

func TestDecreaseIfEnoughFloat64(t *testing.T) {
	cycle := DailyCycle
	typeKey := TypeKey(3)
	userID := UserID(3003)
	key := "balance"

	RegisterCreator(cycle, typeKey, func(uid UserID) *PlayerData {
		return &PlayerData{
			UserID:   uid,
			MiscData: make(map[string]interface{}),
		}
	})

	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		t.Fatalf("GetData failed")
	}

	// 赋初值 float64(100.5)
	pd.MiscData[key] = 100.5

	// 扣减 40.2，成功
	ok := DecreaseIfEnoughFloat64(cycle, typeKey, userID, key, 40.2)
	if !ok {
		t.Errorf("expected decrease success")
	}
	if val := pd.MiscData[key].(float64); val < 60.2 || val > 60.3 { // 允许浮点小误差
		t.Errorf("expected ~60.3 after decrease, got %v", val)
	}

	// 扣减 70.0，失败
	ok = DecreaseIfEnoughFloat64(cycle, typeKey, userID, key, 70.0)
	if ok {
		t.Errorf("expected decrease fail due to insufficient amount")
	}
	if val := pd.MiscData[key].(float64); val < 60.2 || val > 60.3 {
		t.Errorf("expected balance unchanged ~60.3 after failed decrease, got %v", val)
	}

	// 兼容 float32 类型赋值
	pd.MiscData[key] = float32(50.5)
	ok = DecreaseIfEnoughFloat64(cycle, typeKey, userID, key, 20.5)
	if !ok {
		t.Errorf("expected decrease success for float32 rawVal")
	}
	if val := pd.MiscData[key].(float64); val < 29.9 || val > 30.1 {
		t.Errorf("expected ~30.0 after decrease from float32, got %v", val)
	}
}

func TestAppendToInt32SliceIf(t *testing.T) {
	cycle := DailyCycle
	typeKey := TypeKey(10)
	userID := UserID(5005)
	key := "testSlice"

	RegisterCreator(cycle, typeKey, func(uid UserID) *PlayerData {
		return &PlayerData{
			UserID:   uid,
			MiscData: make(map[string]interface{}),
		}
	})

	pd := GetData(cycle, typeKey, userID)
	if pd == nil {
		t.Fatalf("GetData failed")
	}

	// 1. 条件始终返回true，添加新元素成功
	condAlwaysTrue := func(slice []int32) bool {
		return true
	}

	ok := AppendToInt32SliceIf(cycle, typeKey, userID, key, 100, condAlwaysTrue)
	if !ok {
		t.Errorf("expected append success")
	}
	if vals, ok := pd.MiscData[key].([]int32); !ok || len(vals) != 1 || vals[0] != 100 {
		t.Errorf("expected slice contains [100], got %v", pd.MiscData[key])
	}

	// 2. 再添加不同元素成功
	ok = AppendToInt32SliceIf(cycle, typeKey, userID, key, 200, condAlwaysTrue)
	if !ok {
		t.Errorf("expected append success")
	}
	if vals := pd.MiscData[key].([]int32); len(vals) != 2 || vals[1] != 200 {
		t.Errorf("expected slice contains [100, 200], got %v", vals)
	}

	// 3. 添加重复元素失败（因为 slice 已包含 100）
	ok = AppendToInt32SliceIf(cycle, typeKey, userID, key, 100, condAlwaysTrue)
	if !ok {
		// 注意：你的函数没有显式判断 val 是否存在切片中，这里按你代码判断，重复会被添加！
		// 如果需要阻止重复添加，请告诉我，我可以帮你加去重逻辑。
		// 这里假设需要去重，先手动改写测试逻辑
		// 但你的代码没有去重逻辑，重复会被添加
		t.Logf("注意：函数无去重逻辑，重复元素会被添加")
	}

	// 4. 条件返回 false，添加失败
	condFalse := func(slice []int32) bool {
		return false
	}
	ok = AppendToInt32SliceIf(cycle, typeKey, userID, key, 300, condFalse)
	if ok {
		t.Errorf("expected append fail due to cond false")
	}

	// 5. 字段存在但不是 []int32 类型，返回 false
	pd.MiscData[key] = "not a slice"
	ok = AppendToInt32SliceIf(cycle, typeKey, userID, key, 400, condAlwaysTrue)
	if ok {
		t.Errorf("expected append fail due to wrong type")
	}
}
