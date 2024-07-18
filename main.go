package main

import (
	"fmt"
	"sync"
	"time"
)

// Приложение эмулирует получение и обработку неких тасков. Пытается и получать, и обрабатывать в многопоточном режиме.
// Приложение должно генерировать таски 10 сек. Каждые 3 секунды должно выводить в консоль результат всех обработанных к этому моменту тасков (отдельно успешные и отдельно с ошибками).

// ЗАДАНИЕ: сделать из плохого кода хороший и рабочий - as best as you can.
// Важно сохранить логику появления ошибочных тасков.
// Важно оставить асинхронные генерацию и обработку тасков.
// Сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через pull-request в github
// Как видите, никаких привязок к внешним сервисам нет - полный карт-бланш на модификацию кода.

// A Task represents a meaninglessness of our life
type Task struct {
	id         int
	createdAt  string // время создания
	runTime    string // время выполнения
	taskResult []byte
}

// Генерирует таск и отправляет его в каннал
func taskCreator(a chan Task, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	end := time.Now().Add(10 * time.Second)
	for time.Now().Before(end) {
		ft := time.Now().Format(time.RFC3339)
		if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
			ft = "Some error occurred"
		}
		a <- Task{id: int(time.Now().Unix()), createdAt: ft}
	}
	fmt.Println("taskCreator end")
}

func taskWorker(a Task) Task {
	tt, _ := time.Parse(time.RFC3339, a.createdAt)
	if tt.After(time.Now().Add(-20 * time.Second)) {
		a.taskResult = []byte("Task has been succeeded")
	} else {
		a.taskResult = []byte("Something went wrong")
	}
	a.runTime = time.Now().Format(time.RFC3339Nano)
	time.Sleep(time.Millisecond * 150)
	return a
}

func taskSorter(t Task, doneTasks chan Task, undoneTasks chan error) {
	if string(t.taskResult) == "Task has been succeeded" && t.createdAt != "Some error occurred" {
		doneTasks <- t
	} else {
		undoneTasks <- fmt.Errorf("Task id %d time %s, error %s", t.id, t.createdAt, t.taskResult)
	}
}

func printResults(doneTasks []Task, undoneTasks []error) {
	fmt.Println("Errors:")
	for _, err := range undoneTasks {
		fmt.Println(err)
	}

	fmt.Println("Done tasks:")
	for _, task := range doneTasks {
		fmt.Printf("Task id: %d, Created at: %s, Completed at: %s, Result: %s\n",
			task.id, task.createdAt, task.runTime, task.taskResult)
	}
}

func main() {
	var wg sync.WaitGroup
	superChan := make(chan Task, 10)

	go taskCreator(superChan, &wg)

	doneTasks := make(chan Task, 10)
	undoneTasks := make(chan error, 10)
	go func() {
		for t := range superChan {
			var tsk = taskWorker(t)
			go taskSorter(tsk, doneTasks, undoneTasks)
		}
	}()

	// слайсы для хранения тасков
	var okTask []Task
	var errTask []error

	go func() {
		for task := range undoneTasks {
			errTask = append(errTask, task)
		}
	}()

	go func() {
		for task := range doneTasks {
			okTask = append(okTask, task)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second * 3)
			printResults(okTask, errTask)
		}
	}()

	wg.Wait()
}
