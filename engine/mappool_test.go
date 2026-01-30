package engine

import (
	"testing"
)

// TestMapPoolCleanliness 测试对象池中的对象是否干净
func TestMapPoolCleanliness(t *testing.T) {
	// 测试 map[any]bool 对象池
	t.Run("MapAnyBool", func(t *testing.T) {
		// 获取一个对象
		m := GetMap()
		if len(m) != 0 {
			t.Errorf("GetMap() 返回的 map 不为空，长度为 %d", len(m))
		}

		// 使用对象
		m["key"] = true
		if len(m) != 1 {
			t.Errorf("使用后 map 长度应为 1，实际为 %d", len(m))
		}

		// 归还对象
		PutMap(m)

		// 再次获取对象，应该是空的
		m2 := GetMap()
		if len(m2) != 0 {
			t.Errorf("再次 GetMap() 返回的 map 不为空，长度为 %d", len(m2))
		}
		PutMap(m2)
	})

	// 测试 []string 对象池
	t.Run("StringSlice", func(t *testing.T) {
		// 获取一个对象
		s := GetStringSlice()
		if len(s) != 0 {
			t.Errorf("GetStringSlice() 返回的 slice 不为空，长度为 %d", len(s))
		}

		// 使用对象
		s = append(s, "value")
		if len(s) != 1 {
			t.Errorf("使用后 slice 长度应为 1，实际为 %d", len(s))
		}

		// 归还对象
		PutStringSlice(s)

		// 再次获取对象，应该是空的
		s2 := GetStringSlice()
		if len(s2) != 0 {
			t.Errorf("再次 GetStringSlice() 返回的 slice 不为空，长度为 %d", len(s2))
		}
		PutStringSlice(s2)
	})

	// 测试 map[string]any 对象池

	// 测试多次 Get 和 Put 操作
	t.Run("MultipleOperations", func(t *testing.T) {
		// 测试 100 次 Get 和 Put 操作
		for i := 0; i < 100; i++ {
			// 测试 map[any]bool
			m1 := GetMap()
			if len(m1) != 0 {
				t.Errorf("第 %d 次 GetMap() 返回的 map 不为空，长度为 %d", i, len(m1))
			}
			m1["key"] = true
			PutMap(m1)

			// 测试 map[string][]byte

			// 测试 []string
			s := GetStringSlice()
			if len(s) != 0 {
				t.Errorf("第 %d 次 GetStringSlice() 返回的 slice 不为空，长度为 %d", i, len(s))
			}
			s = append(s, "value")
			PutStringSlice(s)

		}
	})
}

// TestMapCreationStrategyCleanliness 测试分类策略返回的对象是否干净
func TestMapCreationStrategyCleanliness(t *testing.T) {
	// 测试 stringSlice 策略
	t.Run("StringSliceStrategy", func(t *testing.T) {
		s := GetStringSliceWithStrategy()
		if len(s) != 0 {
			t.Errorf("GetStringSliceWithStrategy() 返回的 slice 不为空，长度为 %d", len(s))
		}
		PutStringSliceWithStrategy(s)
	})

	// 测试 stringBytesMap 策略
	t.Run("StringBytesMapStrategy", func(t *testing.T) {
		m := GetStringBytesMapWithStrategy()
		if len(m) != 0 {
			t.Errorf("GetStringBytesMapWithStrategy() 返回的 map 不为空，长度为 %d", len(m))
		}
	})

	// 测试 stringAnyMap 策略
	t.Run("StringAnyMapStrategy", func(t *testing.T) {
		m := GetStringAnyMapWithStrategy()
		if len(m) != 0 {
			t.Errorf("GetStringAnyMapWithStrategy() 返回的 map 不为空，长度为 %d", len(m))
		}
	})

	// 测试 anyBoolMap 策略
	t.Run("AnyBoolMapStrategy", func(t *testing.T) {
		m := GetAnyBoolMapWithStrategy()
		if len(m) != 0 {
			t.Errorf("GetAnyBoolMapWithStrategy() 返回的 map 不为空，长度为 %d", len(m))
		}
	})
}
