package util

import (
	"fmt"
	"sync"
)

type Queue struct {
	queue []Job
	lock  sync.Mutex
}

func CreateQueue() *Queue {
	return &Queue{}
}

func (q *Queue) Enqueue(data Job) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.queue = append(q.queue, data)
}

func (q *Queue) Dequeue() Job {
	q.lock.Lock()
	defer q.lock.Unlock()
	var val Job = q.queue[0]
	q.queue = q.queue[1:]
	return val
}

func (q *Queue) Peek() any {
	if q.IsEmpty() {
		return ""
	}
	return q.queue[0]
}

func (q *Queue) IsEmpty() bool {
	return len(q.queue) <= 0
}

func (q *Queue) Display() {
	for i := 0; i < len(q.queue); i++ {
		fmt.Println(q.queue[i])
	}
}
