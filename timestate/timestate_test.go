package timestate

import (
	"context"
	"testing"
	"time"
)

// 修改测试函数签名为 test *testing.T
func TestInitTimezoneTimerAndBasics(test *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		test.Fatalf("加载时区失败: %v", err)
	}

	InitTimezoneTimer(loc)

	// 等待初始化完成，读取零点时间戳
	zeroMs := GetZeroMs()
	zeroSec := GetZeroSec()
	zeroDate := GetZeroDate()

	if zeroMs == 0 || zeroSec == 0 || zeroDate == "" {
		test.Errorf("初始化零点时间戳失败 zeroMs=%d zeroSec=%d zeroDate=%s", zeroMs, zeroSec, zeroDate)
	}

	// 测试周期Key生成
	now := time.Now().In(loc)
	periodKey := GetCurrentPeriodKey()
	expectedKey := BuildPeriodKey(now)
	if periodKey != expectedKey {
		test.Errorf("周期 Key 错误，期望 %s，得到 %s", expectedKey, periodKey)
	}

	// 测试下一个周期 Key
	nextPeriod := GetNextPeriodKey()
	expectedNext := BuildPeriodKey(now.AddDate(0, 0, 7))
	if nextPeriod != expectedNext {
		test.Errorf("下一周期 Key 错误，期望 %s，得到 %s", expectedNext, nextPeriod)
	}
}

func TestGetWeekOfMonth(t *testing.T) {
	testCases := []struct {
		date     time.Time
		expected int
	}{
		{time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC), 1},  // 5月1日 周四，应该是第1周
		{time.Date(2025, 5, 4, 0, 0, 0, 0, time.UTC), 1},  // 5月4日 周日，还是第1周（周一起算）
		{time.Date(2025, 5, 5, 0, 0, 0, 0, time.UTC), 2},  // 5月5日 周一，开始第2周
		{time.Date(2025, 5, 7, 0, 0, 0, 0, time.UTC), 2},  // 5月7日 周三，第2周
		{time.Date(2025, 5, 12, 0, 0, 0, 0, time.UTC), 3}, // 5月12日 周一，第2周
		{time.Date(2025, 5, 13, 0, 0, 0, 0, time.UTC), 3}, // 5月13日 周二，第3周
	}

	for _, tc := range testCases {
		got := GetWeekOfMonth(tc.date)
		if got != tc.expected {
			t.Errorf("GetWeekOfMonth(%v) = %d; want %d", tc.date, got, tc.expected)
		}
	}
}

func TestDayWeekMonthCallbacks(test *testing.T) {
	done := make(chan struct{})
	// 清理旧注册，避免多次测试时重复执行
	RegisterDayCallback(func(t time.Time) { close(done) })

	// 手动触发，观察是否调用
	dayCallbacks.triggerCallbacks(time.Now())

	select {
	case <-done:
		// 正常
	case <-time.After(time.Second):
		test.Error("DayCallback未被触发")
	}

	// 测试周回调
	doneWeek := make(chan struct{})
	RegisterWeekCallback(func(t time.Time) { close(doneWeek) })
	weekCallbacks.triggerCallbacks(time.Now())

	select {
	case <-doneWeek:
	case <-time.After(time.Second):
		test.Error("WeekCallback未被触发")
	}

	// 测试月回调
	doneMonth := make(chan struct{})
	RegisterMonthCallback(func(t time.Time) { close(doneMonth) })
	monthCallbacks.triggerCallbacks(time.Now())

	select {
	case <-doneMonth:
	case <-time.After(time.Second):
		test.Error("MonthCallback未被触发")
	}
}

func TestRegisterDailyTimeCallbackAndTrigger(test *testing.T) {
	// 测试注册无效时间
	RegisterDailyTimeCallback(-1, 0, func(ctx context.Context, t time.Time) {
		test.Error("不应该注册无效时间的回调")
	})
	RegisterDailyTimeCallback(0, 60, func(ctx context.Context, t time.Time) {
		test.Error("不应该注册无效时间的回调")
	})

	called := make(chan struct{})
	RegisterDailyTimeCallback(23, 59, func(ctx context.Context, t time.Time) {
		close(called)
	})

	// 触发刚注册的回调，带超时 3 秒
	TriggerDailyCallbackWithTimeout(23, 59, 3*time.Second)

	select {
	case <-called:
	case <-time.After(time.Second):
		test.Error("每日定时回调未被触发")
	}
}

func TestTriggerDailyCallbackTimeout(test *testing.T) {
	// 注册一个超时回调
	RegisterDailyTimeCallback(1, 1, func(ctx context.Context, t time.Time) {
		<-ctx.Done() // 模拟阻塞直到超时
	})

	// 触发回调，设置超时 1 秒
	TriggerDailyCallbackWithTimeout(1, 1, 1*time.Second)
	// 能走到这里表示触发过程没有死锁或崩溃
}

func TestReadFunctions(test *testing.T) {
	// 读取时间戳相关接口
	if GetZeroMs() == 0 {
		test.Error("GetZeroMs 返回 0")
	}
	if GetZeroSec() == 0 {
		test.Error("GetZeroSec 返回 0")
	}
	if GetZeroDate() == "" {
		test.Error("GetZeroDate 返回空字符串")
	}
	if GetZeroMonth() == "" {
		test.Error("GetZeroMonth 返回空字符串")
	}
	if GetCurrentPeriodKey() == "" {
		test.Error("GetCurrentPeriodKey 返回空字符串")
	}
	if GetNextPeriodKey() == "" {
		test.Error("GetNextPeriodKey 返回空字符串")
	}
	if GetNextDayTimestamp() == 0 {
		test.Error("GetNextDayTimestamp 返回 0")
	}
	if GetNextWeekTimestamp() == 0 {
		test.Error("GetNextWeekTimestamp 返回 0")
	}
	if GetNextMonthTimestamp() == 0 {
		test.Error("GetNextMonthTimestamp 返回 0")
	}
}

func TestMsSecTimer(test *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		test.Fatalf("加载时区失败: %v", err)
	}
	// 初始化定时器
	InitMsSecTimer(loc)

	// 等待定时器启动稳定
	time.Sleep(150 * time.Millisecond)

	ms1 := GetCurrentMs()
	sec1 := GetCurrentSec()

	// 等待超过 1 秒，让定时器更新
	time.Sleep(1100 * time.Millisecond)

	ms2 := GetCurrentMs()
	sec2 := GetCurrentSec()

	// ms 时间戳应当递增或至少不小于之前的值
	if ms2 < ms1 {
		test.Errorf("GetCurrentMs 时间戳不递增：%d < %d", ms2, ms1)
	}

	// sec 时间戳应当至少增加1秒
	if sec2 < sec1+1 {
		test.Errorf("GetCurrentSec 时间戳未增加至少1秒：%d < %d+1", sec2, sec1)
	}

	// 测试 GetTimeSnapshot 返回的值是否一致
	snapshot := GetTimeSnapshot()
	if snapshot.Ms != GetCurrentMs() {
		test.Errorf("GetTimeSnapshot Ms 值不一致，snapshot: %d, GetCurrentMs: %d", snapshot.Ms, GetCurrentMs())
	}
	if snapshot.Sec != GetCurrentSec() {
		test.Errorf("GetTimeSnapshot Sec 值不一致，snapshot: %d, GetCurrentSec: %d", snapshot.Sec, GetCurrentSec())
	}

	// 测试时间戳是否在合理范围内（与系统时间比较，允许 2 秒误差）
	now := time.Now()
	if snapshot.Ms < now.UnixMilli()-2000 || snapshot.Ms > now.UnixMilli()+2000 {
		test.Errorf("GetTimeSnapshot Ms 时间偏差过大: %d vs %d", snapshot.Ms, now.UnixMilli())
	}
	if snapshot.Sec < now.Unix()-2 || snapshot.Sec > now.Unix()+2 {
		test.Errorf("GetTimeSnapshot Sec 时间偏差过大: %d vs %d", snapshot.Sec, now.Unix())
	}
}
