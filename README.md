## What is LazyDB

LazyDB is an innovative Bitcask-based embedded NoSQL database that excels in addressing the limitations of traditional databases, particularly when de
aling with high write-throughput workloads.

## Design of LazyDB

![](https://github.com/billsjc123/LazyDB/blob/main/imgs/architecture.png)

In LazyDB, there are two types of data that needed to be stored either in memory or on disk. The first one is the
actual log data which records the detailed log information. Log data nee
ds to be stored on disk so that it will not lose when you close the databas
e instance. The other one is the index data which is stored in memory.


The diagram below shows the specific design of our database system. It is c
omposed of four layers, API Layer, System Management Layer, In-Memory Stora
ge Layer and Log Storage Layer. Among them, the log storage layer and in-me
mory storage layer as well as garbage collection mechanism are designed in
line with Bitcask's ideas.

![](https://github.com/billsjc123/LazyDB/blob/main/imgs/specific.png)

<details>
    <summary><b>Log storage layer</b></summary>
    All data modification operations are appended to the log and stored in order. This layer ensures data persistence and consistency and can be recovered by re-executing the operations in the log in case of a crash.
</details>

<details>
    <summary><b>In-memory storage layer</b></summary>
    LazyDB uses various in-memory data structures to improve read and write performance. LazyDB also provides two types of indexes, which are Concurrent HashMap and Adaptive Radix Tree. Besides, LazyDB provides a switch for the in-memory database.
</details>    

<details>
    <summary><b>System management layer</b></summary>
    This layer provides the most basic ability to add, delete, change, and query the entire database through the interface of the logging and indexing layers. 
</details>

<details>
    <summary><b>API layer</b></summary>
     LazyDB is providing an interface to external database operations through APIs for five data types. As an embedded database, the user can call these interfaces directly in the code to manipulate the local database.
</details>

## Gettings Started

### Basic operations

```go
package main

import (
	"fmt"
	"github.com/billsjc123/LazyDB"
	"os"
	"path/filepath"
)

func main() {
	// empty db directory
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "tmp")

	// use default config
	cfg := lazydb.DefaultDBConfig(path)

	// open lazydb
	db, err := lazydb.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = db.Close()
	}()

	// string: set a key and value
	if err = db.Set([]byte("str_key"), []byte("str_value")); err != nil {
		panic(err)
	}

	// string: get the of certain key
	value, err := db.Get([]byte("str_key"))
	if err != nil {
		panic(err)
	}
	fmt.Println(string(value))

	// delete a key
	if err = db.Delete([]byte("str_key")); err != nil {
		panic(err)
	}
}
```

### Transactions

```go
	// create a transaction
	tx, err := db.Begin(lazydb.RWTX)
	if err != nil {
		panic(err)
	}
	tx.Set([]byte("1"), []byte("val1"))
	tx.SAdd([]byte("add2"), [][]byte{[]byte("v1")}...)
	tx.Set([]byte("3"), []byte("val3"))

	// commit a transaction
	if err = tx.Commit(); err != nil {
		panic(err)
	}

	val, err := db.Get([]byte("1"))
	if err != nil {
		panic(err)
	}
	println(string(val))
	got := db.SIsMember([]byte("add2"), []byte("v1"))
	println(got)
	val, err = db.Get([]byte("3"))
	if err != nil {
		panic(err)
	}
	println(string(val))

