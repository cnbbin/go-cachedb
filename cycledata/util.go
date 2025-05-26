package cycledata


func toInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int32:
		return int(val), true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	default:
		return 0, false
	}
}

func toInt32(v interface{}) (int32, bool) {
	switch val := v.(type) {
	case int32:
		return val, true
	case int:
		return int32(val), true
	case int64:
		return int32(val), true
	case float64:
		return int32(val), true
	default:
		return 0, false
	}
}

func toInt64(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case int64:
		return val, true
	case int:
		return int64(val), true
	case int32:
		return int64(val), true
	case float64:
		return int64(val), true
	default:
		return 0, false
	}
}

// 工具函数：将 interface{} 转换为 float64（支持 int, int64, float32, float64）
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	default:
		return 0, false
	}
}

func toInt32Slice(v interface{}) ([]int32, bool) {
	switch s := v.(type) {
	case []int32:
		return s, true
	case []interface{}:
		result := make([]int32, 0, len(s))
		for _, item := range s {
			switch n := item.(type) {
			case int32:
				result = append(result, n)
			case int:
				result = append(result, int32(n))
			case int64:
				result = append(result, int32(n))
			case float64:
				result = append(result, int32(n))
			default:
				return nil, false
			}
		}
		return result, true
	default:
		return nil, false
	}
}

func UtilCopyMap(src map[string]interface{}) map[string]interface{} {
    dst := make(map[string]interface{} , len(src))
    for k, v := range src {
        dst[k] = v
    }
    return dst
}