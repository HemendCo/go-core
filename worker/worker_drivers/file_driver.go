package worker_drivers

import (
	"encoding/json"
	"fmt"
	"github.com/HemendCo/go-core"
	"github.com/HemendCo/go-core/filemanager"
	"github.com/HemendCo/go-core/helpers"
	"github.com/HemendCo/go-core/worker/worker_interfaces"
	"github.com/HemendCo/go-core/worker/worker_models"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// FileWorkerDriver represents a worker driver that manages tasks using a file.
type FileWorkerDriver struct {
	app         *core.App
	cfg         worker_models.FileWorkerConfig
	path        string                   // Path to the file used for storing tasks
	fileManager *filemanager.FileManager // FileManager for file operations
	jobs        []worker_interfaces.Job  // Registered job handlers
	location    *time.Location
}

// Name returns the name of the worker driver.
func (f *FileWorkerDriver) Name() string {
	return "file"
}

// Init initializes the worker driver with the application and configuration.
func (f *FileWorkerDriver) Init(app *core.App, config interface{}) error {
	cfg, ok := config.(worker_models.FileWorkerConfig)
	if !ok {
		return fmt.Errorf("unsupported worker config: worker_models.FileWorkerConfig")
	}

	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return err
	}

	f.app = app
	f.cfg = cfg
	f.fileManager = filemanager.NewFileManager() // Initialize FileManager
	f.path = helpers.JoinWithProjectPath(f.cfg.Path)
	f.location = location

	return os.MkdirAll(f.path, os.ModePerm) // Ensure the file exists
}

// Enqueue adds a new job to the file queue.
func (f *FileWorkerDriver) Enqueue(job worker_interfaces.Job, params interface{}) error {
	preJson, err := job.NewTask(f.app, params)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(preJson)
	if err != nil {
		return err
	}

	taskID := uuid.New().String()

	currentTime := time.Now().In(f.location)
	formattedTime := currentTime.Format("2006-01-02T15-04-05")

	fileName := fmt.Sprintf("%s_%s", formattedTime, taskID)

	task := worker_models.TaskMessage{
		ID:        taskID,
		TypeName:  job.TypeName(),
		FileName:  fileName,
		Payload:   payload,
		Queue:     "default",
		Retried:   0,
		CreatedAt: currentTime,
	}

	// Save updated tasks to file
	if err := f.fileManager.WriteFile(filepath.Join(f.path, fileName), task); err != nil {
		return fmt.Errorf("failed to write updated tasks: %w", err)
	}

	return nil
}

// RegisterJobHandlers adds multiple job handlers to the worker driver.
func (f *FileWorkerDriver) RegisterJobHandlers(handlers ...worker_interfaces.Job) {
	if len(handlers) == 0 {
		fmt.Println("No job handlers provided to register.")
		return
	}

	for _, handler := range handlers {
		if !f.JobHandlerExists(handler) {
			f.jobs = append(f.jobs, handler)
		} else {
			fmt.Printf("Job handler %s is already registered and will not be added again.\n", handler.TypeName())
		}
	}
}

// JobHandlerExists checks if a specific job handler is already registered.
func (f *FileWorkerDriver) JobHandlerExists(handler worker_interfaces.Job) bool {
	for _, h := range f.jobs {
		if h.TypeName() == handler.TypeName() {
			return true
		}
	}
	return false
}

// Close performs any necessary cleanup for the FileWorkerDriver.
func (f *FileWorkerDriver) Close() error {
	// Any cleanup logic can be added here if needed
	return nil
}

// Run starts processing the tasks with registered job handlers.
func (f *FileWorkerDriver) Run(handlers ...worker_interfaces.Job) error {
	f.RegisterJobHandlers(handlers...)

	timeAndUUIDPattern := `^([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2})_([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})$`
	regex, err := regexp.Compile(timeAndUUIDPattern)
	if err != nil {
		return fmt.Errorf("error compiling regex: %w", err)
	}

	excludePattern := `^\..*`
	regex2, err := regexp.Compile(excludePattern)
	if err != nil {
		return fmt.Errorf("error compiling regex: %w", err)
	}

	for {
		if err := f.processFiles(regex, regex2); err != nil {
			fmt.Printf("Error processing files: %v. Please wait a moment while we attempt the next check...\n", err)
		}
		if f.cfg.CheckInterval > 0 {
			time.Sleep(time.Duration(f.cfg.CheckInterval) * time.Second) // Sleep before next check
		}
	}
}
func (f *FileWorkerDriver) processFiles(regex, regex2 *regexp.Regexp) error {
	files, err := f.fileManager.ListFilesInDirectory(f.path, func(entry os.DirEntry) bool {
		return !regex.MatchString(entry.Name()) || regex2.MatchString(entry.Name())
	})
	if err != nil {
		return fmt.Errorf("error listing files: %w", err)
	}

	log.Printf("Successfully read %d tasks from file. Next check will occur in %d seconds.", len(files), f.cfg.CheckInterval)

	// Create channels for different priority queues based on configuration
	tasks := make(map[string][]worker_models.TaskMessage)
	for priority := range f.cfg.Priorities {
		tasks[priority] = make([]worker_models.TaskMessage, 0) // Initialize slice for each priority
	}

	// Read files and populate tasks slice based on their queue
	for _, entry := range files {
		var task worker_models.TaskMessage
		if err := f.fileManager.ReadFile(filepath.Join(f.path, entry.Name()), &task); err != nil {
			fmt.Printf("Error reading file %s: %v\n", entry.Name(), err)
			continue
		}

		// Use task.Queue to add task to the appropriate slice
		if _, exists := tasks[task.Queue]; exists {
			if len(tasks[task.Queue]) < f.cfg.Priorities[task.Queue] {
				tasks[task.Queue] = append(tasks[task.Queue], task)
			}
		} else {
			fmt.Printf("Unknown queue %s for task %s\n", task.Queue, task.ID)
		}
	}

	var wg sync.WaitGroup

	// Start worker pools based on priority
	for priority := range f.cfg.Priorities {
		wg.Add(1)

		// Send a slice of tasks to a new worker
		go func(tasksToProcess []worker_models.TaskMessage) {
			defer wg.Done()
			for _, task := range tasksToProcess {
				f.executeTask(task)
				if f.cfg.TaskSleepDuration > 0 {
					time.Sleep(time.Duration(f.cfg.TaskSleepDuration * float64(time.Second)))
				}
			}
		}(tasks[priority])
	}

	wg.Wait() // Wait for all workers to finish
	return nil
}

func (f *FileWorkerDriver) executeTask(task worker_models.TaskMessage) {
	for _, handler := range f.jobs {
		if handler.TypeName() == task.TypeName {
			filePath := filepath.Join(f.path, task.FileName)
			payload := []byte(task.Payload)

			if err := handler.Handler(f.app, payload); err != nil {
				fmt.Printf("Error executing task %s: %v\n", task.ID, err)
				task.Retried++

				// Ensure UpdatedAt is initialized if it's a pointer
				if task.UpdatedAt == nil {
					task.UpdatedAt = new(time.Time) // Initialize the pointer if it's nil
				}
				*task.UpdatedAt = time.Now().In(f.location)

				if task.Retried > f.cfg.MaxRetry {
					if err := f.fileManager.DeleteFileIfExists(filePath); err != nil {
						fmt.Printf("Error deleting task file %s: %v\n", task.ID, err)
					} else {
						fmt.Printf("Task %s retried more than %d times and file deleted.\n", task.ID, f.cfg.MaxRetry)
					}
				} else {
					if err := f.fileManager.WriteFile(filePath, task); err != nil {
						fmt.Printf("Error updating task file %s: %v\n", task.ID, err)
					} else {
						fmt.Printf("Task %s retried %d times and file updated.\n", task.ID, task.Retried)
					}
				}
			} else {
				if err := f.fileManager.DeleteFileIfExists(filePath); err != nil {
					fmt.Printf("Error deleting task file %s: %v\n", task.ID, err)
				} else {
					fmt.Printf("Task %s executed successfully and file deleted.\n", task.ID)
				}
			}
			break // exit loop if handler is found
		}
	}
}
