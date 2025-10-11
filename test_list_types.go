package main

import "fmt"

// 测试convertValueByType函数的列表类型支持
func testListTypes() {
	// 创建HTTPTool实例用于测试
	tool := &HTTPTool{}
	
	fmt.Println("=== 测试列表类型转换 ===")
	
	// 测试string[]类型
	fmt.Println("\n1. 测试 string[] 类型:")
	
	// 测试逗号分隔
	result, err := tool.convertValueByType("apple,banana,orange", "string[]")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		strArr := result.([]string)
		fmt.Printf("逗号分隔: %v\n", strArr)
	}
	
	// 测试分号分隔
	result, err = tool.convertValueByType("red;blue;green", "string[]")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		strArr := result.([]string)
		fmt.Printf("分号分隔: %v\n", strArr)
	}
	
	// 测试空格分隔
	result, err = tool.convertValueByType("a b c", "string[]")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		strArr := result.([]string)
		fmt.Printf("空格分隔: %v\n", strArr)
	}
	
	// 测试空值
	result, err = tool.convertValueByType("", "string[]")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		strArr := result.([]string)
		fmt.Printf("空值: %v\n", strArr)
	}
	
	// 测试int[]类型
	fmt.Println("\n2. 测试 int[] 类型:")
	
	// 测试逗号分隔
	result, err = tool.convertValueByType("1,2,3,4,5", "int[]")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		intArr := result.([]int)
		fmt.Printf("逗号分隔: %v\n", intArr)
	}
	
	// 测试分号分隔
	result, err = tool.convertValueByType("10;20;30", "int[]")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		intArr := result.([]int)
		fmt.Printf("分号分隔: %v\n", intArr)
	}
	
	// 测试空值
	result, err = tool.convertValueByType("", "int[]")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		intArr := result.([]int)
		fmt.Printf("空值: %v\n", intArr)
	}
	
	// 测试混合空格和无效值
	result, err = tool.convertValueByType("1, abc, 3, 4.5, 5", "int[]")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		intArr := result.([]int)
		fmt.Printf("混合值: %v\n", intArr)
	}
	
	fmt.Println("\n=== 所有测试完成 ===")
}

// 运行测试
func main() {
	testListTypes()
}