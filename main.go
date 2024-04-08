package main

import (
	"fmt"
	"time"
)

//Можно использовать более понятные имена переменных и типов.
//Например, вместо Ttype использовать Task или подобное название, которое лучше отражает сущность.

//Оптимизация многопоточности:
//Вместо создания новой горутины для каждой задачи или ошибки, можно обрабатывать задачи в пуле горутин,
//что сокращает накладные расходы на управление большим количеством горутин и облегчает контроль над их выполнением.

//Обработка ошибок:
//Можно использовать стандартный паттерн Go для обработки ошибок,
//чтобы сделать код более понятным и безопасным.

//Логика появления ошибочных задач:
//Поддерживать логику появления ошибочных задач можно,
//добавив проверку внутри task_worker или после создания задачи в taskCreturer.

type Task struct {
	id         int64
	createTime string
	finishTime string
	result     []byte
}

func main() {
	taskCreator := func(taskChan chan<- Task) {
		for {
			ct := time.Now()
			ctString := ct.Format(time.RFC3339)
			// Имитировать состояние ошибки
			if ct.Nanosecond()%2 > 0 {
				ctString = "Some error occurred"
			}
			taskChan <- Task{createTime: ctString, id: ct.Unix()}
			time.Sleep(10 * time.Millisecond)
		}
	}

	taskChannel := make(chan Task, 10)
	doneTasks := make(chan Task)
	undoneTasks := make(chan error)

	// Процессор обработки задач
	taskProcessor := func() {
		for task := range taskChannel {
			ct, err := time.Parse(time.RFC3339, task.createTime)
			if err != nil {
				task.result = []byte(fmt.Sprintf("Invalid creation time: %s", err))
				undoneTasks <- fmt.Errorf("task id %d, error: %s", task.id, task.result)
				continue
			}
			if time.Since(ct) < 20*time.Second {
				task.result = []byte("task has been succeeded")
				doneTasks <- task
			} else {
				task.result = []byte("task has failed: time exceeded")
				undoneTasks <- fmt.Errorf("task id %d, error: %s", task.id, task.result)
			}
			task.finishTime = time.Now().Format(time.RFC3339Nano)
		}
	}

	go taskCreator(taskChannel)

	// Запустите фиксированное количество горутины для обработки задач
	for i := 0; i < 4; i++ {
		go taskProcessor()
	}

	// Закрыть каналы по завершении обработки (возможно, потребуется более эффективный контроль в реальном приложении).
	go func() {
		time.Sleep(3 * time.Second) // Имитировать рабочее время
		close(taskChannel)
		close(doneTasks)
		close(undoneTasks)
	}()

	// Вывод результатов
	for task := range doneTasks {
		fmt.Printf("Task completed: %+v\n", task)
	}
	for err := range undoneTasks {
		fmt.Printf("Task error: %v\n", err)
	}
}
