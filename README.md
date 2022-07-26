# Go (Golang) Concurrency Patterns

## Introduction

### What is concurrency?
* Concurrency is the composition of independently executing computations.
* Concurrency is a way to structure software, particularly as a way to write clean code that interacts well with the real world.
* It is not parallelism.

### Concurrency is not parallelism
* Concurrency is not parallelism, although it enables parallelism.
* If you have only one processor, your program can still be concurrent but it cannot be parallel.
* On the other hand, a well-written concurrent program might run efficiently in parallel on a multiprocessor.

### What is a goroutine?
* It's an independently executing function, launched by a go statement.
* It has its own call stack, which grows and shrinks as required.
* It's very cheap. It's practical to have thousands, even hundreds of thousands of goroutines.
* It's not a thread.
* There might be only one thread in a program with thousands of goroutines.
* Instead, goroutines are multiplexed dynamically onto threads as needed to keep all the goroutines running.
* But if you think of it as a very cheap thread, you won't be far off.

_*** Disclaimer: All the definitions used in texts & code are originating from Rob Pike's (one of three original architects & authors of Go language) talk presented at Google I/O in June 2012. [Watch the talk on YouTube](http://www.youtube.com/watch?v=f6kdp27TYZs)_
