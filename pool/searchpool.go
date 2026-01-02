package pool

import (
	"context"
	"sync"
	"time"
)

// SearchTask 定义搜索任务
// Query: 搜索查询词
// Result: 搜索结果（协程安全，由任务池填充）
// Error: 搜索过程中的错误（协程安全，由任务池填充）
type SearchTask struct {
	Query  string      // 搜索查询词
	Result interface{} // 搜索结果
	Error  error       // 搜索错误
}

// SearchFunc 定义搜索函数类型
// 参数：查询词
// 返回：搜索结果和可能的错误
type SearchFunc func(query string) (interface{}, error)

// SearchPool 搜索任务池
// 用于并发执行搜索任务并收集结果
type SearchPool struct {
	size       int           // 工作协程数量
	taskCh     chan *SearchTask // 任务通道
	resultCh   chan *SearchTask // 结果通道
	wg         sync.WaitGroup // 等待组，用于等待所有任务完成
	ctx        context.Context // 上下文，用于控制任务取消
	cancel     context.CancelFunc // 取消函数，用于关闭任务池
	searchFunc SearchFunc // 搜索函数
}

// NewSearchPool 创建一个新的搜索任务池
// size: 工作协程数量，建议设置为CPU核心数或根据实际需求调整
// searchFunc: 实际的搜索函数
// 返回值: 搜索任务池实例
func NewSearchPool(size int, searchFunc SearchFunc) *SearchPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 确保size至少为1
	if size <= 0 {
		size = 1
	}
	
	return &SearchPool{
		size:       size,
		taskCh:     make(chan *SearchTask, 100), // 带缓冲的任务通道
		resultCh:   make(chan *SearchTask, 100), // 带缓冲的结果通道
		ctx:        ctx,
		cancel:     cancel,
		searchFunc: searchFunc,
	}
}

// Start 启动搜索任务池的工作协程
func (p *SearchPool) Start() {
	// 启动指定数量的工作协程
	for i := 0; i < p.size; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker 工作协程，执行搜索任务
func (p *SearchPool) worker() {
	defer p.wg.Done()
	
	for {
		select {
		case <-p.ctx.Done():
			// 任务池已关闭，退出
			return
		case task, ok := <-p.taskCh:
			if !ok {
				// 任务通道已关闭，退出
				return
			}
			
			// 执行搜索任务
			result, err := p.searchFunc(task.Query)
			task.Result = result
			task.Error = err
			
			// 将结果发送到结果通道
			p.resultCh <- task
		}
	}
}

// collector 结果收集协程，等待所有任务完成后关闭结果通道
// 注意：这个方法不再通过 Start 启动，而是在 GetResults 中调用
func (p *SearchPool) collector() {
	defer close(p.resultCh)
	
	// 等待所有工作协程完成
	p.wg.Wait()
}

// Submit 提交搜索任务
// task: 搜索任务指针
// 返回值: 如果任务池已关闭，返回错误；否则返回nil
func (p *SearchPool) Submit(task *SearchTask) error {
	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.taskCh <- task:
		return nil
	}
}

// SubmitMany 批量提交搜索任务
// tasks: 搜索任务切片
// 返回值: 如果任务池已关闭，返回错误；否则返回nil
func (p *SearchPool) SubmitMany(tasks []*SearchTask) error {
	for _, task := range tasks {
		if err := p.Submit(task); err != nil {
			return err
		}
	}
	return nil
}

// GetResults 获取所有搜索结果
// 返回值: 搜索结果切片和可能的错误
func (p *SearchPool) GetResults() ([]*SearchTask, error) {
	// 关闭任务通道，不再接受新任务
	close(p.taskCh)
	
	// 启动结果收集协程，等待所有工作协程完成后关闭结果通道
	go p.collector()
	
	var results []*SearchTask
	
	// 从结果通道读取所有结果
	for task := range p.resultCh {
		results = append(results, task)
	}
	
	return results, nil
}

// GetResultsWithTimeout 获取所有搜索结果，支持超时
// timeout: 超时时间
// 返回值: 搜索结果切片和可能的错误（包括超时错误）
func (p *SearchPool) GetResultsWithTimeout(timeout time.Duration) ([]*SearchTask, error) {
	// 关闭任务通道，不再接受新任务
	close(p.taskCh)
	
	// 启动结果收集协程，等待所有工作协程完成后关闭结果通道
	go p.collector()
	
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()
	
	var results []*SearchTask
	
	// 使用select监听结果通道和超时
	for {
		select {
		case <-timeoutCtx.Done():
			// 超时
			return results, timeoutCtx.Err()
		case task, ok := <-p.resultCh:
			if !ok {
				// 结果通道已关闭，所有结果已处理完成
				return results, nil
			}
			results = append(results, task)
		}
	}
}

// Close 关闭搜索任务池，释放资源
// 会取消所有未完成的任务
func (p *SearchPool) Close() {
	p.cancel()
}

// Size 返回工作协程数量
func (p *SearchPool) Size() int {
	return p.size
}

// 使用示例
/*
// 示例搜索函数
func exampleSearchFunc(query string) (interface{}, error) {
	// 模拟搜索耗时
	time.Sleep(100 * time.Millisecond)
	
	// 模拟搜索结果
	result := map[string]interface{}{
		"query":   query,
		"result":  "搜索结果: " + query,
		"time":    time.Now().Format(time.RFC3339),
	}
	
	return result, nil
}

// 示例用法
func main() {
	// 创建搜索任务池，使用4个工作协程
	pool := pool.NewSearchPool(4, exampleSearchFunc)
	defer pool.Close()
	
	// 启动任务池
	pool.Start()
	
	// 创建搜索任务
	tasks := []*pool.SearchTask{
		{Query: "golang"},
		{Query: "python"},
		{Query: "java"},
		{Query: "javascript"},
		{Query: "rust"},
	}
	
	// 提交任务
	if err := pool.SubmitMany(tasks); err != nil {
		fmt.Printf("提交任务失败: %v\n", err)
		return
	}
	
	// 获取结果，超时时间1秒
	results, err := pool.GetResultsWithTimeout(time.Second)
	if err != nil {
		fmt.Printf("获取结果失败: %v\n", err)
		return
	}
	
	// 打印结果
	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("查询 '%s' 失败: %v\n", result.Query, result.Error)
		} else {
			fmt.Printf("查询 '%s' 成功: %v\n", result.Query, result.Result)
		}
	}
}
*/
