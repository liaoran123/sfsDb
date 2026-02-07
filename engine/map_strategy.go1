package engine

// MapCreationStrategy 提供基于对象类型的混合创建策略
// 对于不同类型的对象，选择最合适的创建方式

// GetStringSliceWithStrategy 获取 []string 切片
// 对于 []string 类型，使用对象池可以获得显著的性能提升
func GetStringSliceWithStrategy() []string {
	return GetStringSlice() // 使用对象池
}

// PutStringSliceWithStrategy 归还 []string 切片
func PutStringSliceWithStrategy(s []string) {
	PutStringSlice(s) // 归还到对象池
}

// GetStringBytesMapWithStrategy 获取 map[string][]byte 映射
// 对于 map[string][]byte 类型，直接创建性能更好
func GetStringBytesMapWithStrategy() map[string][]byte {
	return make(map[string][]byte) // 直接创建
}

// GetStringAnyMapWithStrategy 获取 map[string]any 映射
// 对于 map[string]any 类型，直接创建性能更好
func GetStringAnyMapWithStrategy() map[string]any {
	return make(map[string]any) // 直接创建
}

// GetAnyBoolMapWithStrategy 获取 map[any]bool 映射
// 对于 map[any]bool 类型，直接创建性能更好
func GetAnyBoolMapWithStrategy() map[any]bool {
	return make(map[any]bool) // 直接创建
}

// MapCreationStrategy 提供统一的对象创建策略接口
// 根据对象类型选择最合适的创建方式
func MapCreationStrategy(mapType string) any {
	switch mapType {
	case "stringSlice":
		return GetStringSliceWithStrategy()
	case "stringBytesMap":
		return GetStringBytesMapWithStrategy()
	case "stringAnyMap":
		return GetStringAnyMapWithStrategy()
	case "anyBoolMap":
		return GetAnyBoolMapWithStrategy()
	default:
		return nil
	}
}
