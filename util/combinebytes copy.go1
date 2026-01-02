package util

//全文索引算法
// CombineBytes 接收多个字节数组，使用指定分隔符返回第一个数组与其他数组的所有可能拼接组合
func CombineBytes(arrays [][][]byte, sep []byte) [][]byte {
	if len(arrays) == 0 {
		return [][]byte{}
	}
	// 从第一个数组开始，复制元素以避免修改原始数据
	result := GetBytesArray() //全文索引数据量大，所以使用对象池，避免频繁分配内存。使用后需要put到对象池
	for _, b := range arrays[0] {
		result = append(result, append([]byte{}, b...)) // 复制数组，避免共享底层数组
	}
	// 处理剩余数组，生成所有可能的拼接组合
	for _, array := range arrays[1:] {
		newResult := make([][]byte, 0, len(result)*len(array))
		for _, existing := range result {
			for _, nextBytes := range array {
				// 创建新的组合：existing + sep + nextBytes
				combined := make([]byte, 0, len(existing)+len(sep)+len(nextBytes))
				combined = append(combined, existing...)
				combined = append(combined, sep...)
				combined = append(combined, nextBytes...)
				newResult = append(newResult, combined)
			}
		}
		result = newResult
	}
	/*
		for _, b := range result {
			fmt.Printf("b: %v\n", string(b))
		}*/
	return result
}

// CombineStrings 接收多个字符串数组，使用指定分隔符返回第一个数组与其他数组的所有可能拼接组合
// 这是 CombineBytes 的字符串包装版本
func CombineStrings(arrays [][]string, sep string) []string {
	// 转换为字节数组
	byteArrays := make([][][]byte, len(arrays))
	for i, array := range arrays {
		byteArrays[i] = make([][]byte, len(array))
		for j, s := range array {
			byteArrays[i][j] = []byte(s)
		}
	}

	// 调用 CombineBytes
	byteResult := CombineBytes(byteArrays, []byte(sep))

	// 转换回字符串数组
	result := make([]string, len(byteResult))
	for i, b := range byteResult {
		result[i] = string(b)
	}

	return result
}
