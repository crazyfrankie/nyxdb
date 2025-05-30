# NyxDB
NyxDB 是一个基于 LSM（Log-Structured Merge Tree）原理设计的分布式键值存储数据库，旨在提供高效的写入性能和稳定的查询响应。
NyxDB 能够横向扩展以满足大规模应用的需求，适用于需要快速写入和大规模数据存储的场景。

## 名字来源
Nyx（纽克斯）是希腊神话中的夜神，象征着神秘和深邃。我们选择这个名字，是因为 NyxDB 专注于为用户提供一个 稳定、高效、可扩展的存储引擎，就像夜晚的深邃与宁静一样，它在后台默默地保障着数据的安全与可靠，确保系统在高负载情况下依旧能够保持高效运转。

## 特性
- 基于 LSM 树：NyxDB 使用 LSM 树（Log-Structured Merge Tree）作为存储引擎，能够在高并发写入场景下提供出色的性能。 
- 分布式架构：支持分布式部署，能够根据需求动态扩展节点，适应大规模数据存储与高吞吐量需求。 
- 高效的写入性能：得益于 LSM 树的设计，NyxDB 在写入密集型应用中表现优异，能够高效地处理大量的写操作。 
- 可靠的数据存储：支持写前日志（WAL）和定期合并（Compaction）机制，确保数据一致性和持久化。 
- 水平扩展性：支持通过一致性哈希分片和自动负载均衡，轻松实现集群横向扩展。 
- 简洁易用：提供简洁的 API 接口，支持常见的键值操作，能够无缝集成到现有应用中。

## 安装与使用
### 安装
你可以通过以下方式安装 NyxDB：

```bash
go get github.com/crazyfrankie/nyxdb
```

### 示例
以下是使用 NyxDB 的简单示例：

```go
package main

import (
    "fmt"
    "log"

	"github.com/crazyfrankie/nyxdb"
)

func main() {
    // 创建一个新的数据库实例
    db, err := nyxdb.New()
    if err != nil {
    log.Fatal(err)
    }

	// 设置键值对
	err = db.Put("key1", "value1")
	if err != nil {
		log.Fatal(err)
	}

	// 获取键对应的值
	value, err := db.Get("key1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("key1 ->", value)

	// 删除键值对
	err = db.Delete("key1")
	if err != nil {
		log.Fatal(err)
	}
}
```

### 配置
NyxDB 支持基本的配置文件设置，你可以通过修改配置文件来调整数据库行为。例如：

```yaml
server:
  port: 8080
  cluster: true
  nodes:
    - 192.168.1.1:9921
    - 192.168.1.2:9921

storage:
  compaction_interval: 3600
  memtable_size: 64MB
```

## 数据持久化
NyxDB 提供了持久化机制，数据会通过写前日志（WAL）和周期性的合并过程（Compaction）来确保数据的可靠存储。
