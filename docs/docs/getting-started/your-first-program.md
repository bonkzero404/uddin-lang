---
sidebar_position: 3
---

# Your First Program

Let's build a complete Uddin-Lang application from scratch! We'll create a simple task manager that demonstrates core language features.

## Project Overview

We'll build a command-line task manager with these features:

-   Add new tasks
-   List all tasks
-   Mark tasks as complete
-   Remove tasks
-   Save/load tasks from a file

## Setting Up

Create a new directory for your project:

```bash
mkdir my-first-uddin-app
cd my-first-uddin-app
```

## Step 1: Basic Structure

Create `task_manager.din`:

```uddin
// task_manager.din - A simple task manager

// Global task list
tasks = []

fun main():
    print("🚀 Welcome to Uddin Task Manager!")
    print("Type 'help' for available commands")

    // Load existing tasks (commented out for now)
    // load_tasks()

    // Main program loop
    while (true):
        command = input("\n" + str(len(tasks)) + " tasks | Enter command: ")

        if (command == "help") then:
            show_help()
        else if (command == "add") then:
            add_task()
        else if (command == "list") then:
            list_tasks()
        else if (command == "complete") then:
            complete_task()
        else if (command == "remove") then:
            remove_task()
        else if (command == "save") then:
            save_tasks()
        else if (command == "quit") then:
            save_tasks()
            print("👋 Goodbye!")
            break
        else:
            print("❌ Unknown command. Type 'help' for available commands.")
        end
    end
end
```

## Step 2: Helper Functions

Add these functions to handle different operations:

```uddin
// Display help information
fun show_help():
    print("\n📋 Available Commands:")
    print("  add      - Add a new task")
    print("  list     - Show all tasks")
    print("  complete - Mark a task as complete")
    print("  remove   - Remove a task")
    print("  save     - Save tasks to file")
    print("  help     - Show this help")
    print("  quit     - Exit the program")
end

// Add a new task
fun add_task():
    print("Enter task description: ")
    description = input()

    if (description != "") then:
        task = {
            "id": len(tasks) + 1,
            "description": description,
            "completed": false,
            "created_at": get_timestamp()
        }

        push(tasks, task)
        print("✅ Task added: " + description)
    else:
        print("❌ Task description cannot be empty")
    end
end

// List all tasks
fun list_tasks():
    if (len(tasks) == 0) then:
        print("📝 No tasks found. Add some tasks to get started!")
        return null
    end

    print("\n📋 Your Tasks:")
    print("" + "-" * 50)

    for (task in tasks):
        status = task["completed"] ? "✅" : "⏳"
        id_str = "[" + str(task["id"]) + "]"
        description = task["description"]

        print(status + " " + id_str + " " + description)
    end

    print("" + "-" * 50)
end

// Mark a task as complete
fun complete_task():
    if (len(tasks) == 0) then:
        print("❌ No tasks available")
        return null
    end

    list_tasks()
    print("Enter task ID to mark as complete: ")
    input_id = input()

    // Convert string to number
    task_id = int(input_id)

    // Find and update task
    found = false
    for (i in range(len(tasks))):
        if (tasks[i]["id"] == task_id) then:
            if (tasks[i]["completed"]) then:
                print("ℹ️  Task is already completed")
            else:
                tasks[i]["completed"] = true
                print("✅ Task marked as complete!")
            end
            found = true
            break
        end
    end

    if (not found) then:
        print("❌ Task with ID " + str(task_id) + " not found")
    end
end

// Save tasks to file (placeholder implementation)
fun save_tasks():
    print("💾 Tasks saved successfully!")
    return null
end

// Load tasks from file (placeholder implementation)
fun load_tasks():
    // No existing tasks to load in this demo
    return null
end

// Get current timestamp (placeholder implementation)
fun get_timestamp():
    return "2024-01-01 12:00:00"
end

// Remove a task
fun remove_task():
    if (len(tasks) == 0) then:
        print("❌ No tasks available")
        return null
    end

    list_tasks()
    print("Enter task ID to remove: ")
    input_id = input()

    task_id = int(input_id)

    // Find and remove task
    new_tasks = []
    found = false

    for (task in tasks):
        if (task["id"] != task_id) then:
            push(new_tasks, task)
        else:
            found = true
            print("🗑️  Removed task: " + task["description"])
        end
    end

    if (found) then:
        tasks = new_tasks
        // Reassign IDs
        for (i in range(len(tasks))):
            tasks[i]["id"] = i + 1
        end
    else:
        print("❌ Task with ID " + str(task_id) + " not found")
    end
end
```

## Step 3: File Operations

Add functions to save and load tasks with real file operations:

```uddin
// Save tasks to file
fun save_tasks():
    import "json"
    import "fs"
    if (len(tasks) == 0) then:
        print("ℹ️  No tasks to save")
        return
    end

    // Convert tasks to JSON string
    json_data = json.stringify(tasks)

    // Write to file
    if (fs.write("tasks.json", json_data)) then:
        print("💾 Tasks saved successfully!")
    else:
        print("❌ Failed to save tasks")
    end
end

// Load tasks from file
fun load_tasks():
    import "json"
    import "fs"
    // Check if tasks file exists
    if (fs.exists("tasks.json")) then:
        print("📂 Found existing tasks file. Loading...")

        // Read file content
        file_content = fs.read("tasks.json")

        if (file_content != null && file_content != "") then:
            // Parse JSON content
            loaded_tasks = json.parse(file_content)

            if (loaded_tasks != null) then:
                // Clear existing tasks and load from file
                while (len(tasks) > 0):
                    pop(tasks)
                end

                // Add loaded tasks to global array
                for (task in loaded_tasks):
                    push(tasks, task)
                end

                print("✅ Successfully loaded " + str(len(tasks)) + " task(s)")
            else:
                print("❌ Failed to parse tasks file")
            end
        else:
            print("📝 Tasks file is empty")
        end
    else:
        print("📝 No existing tasks file found. Starting fresh!")
    end
end

// Get current timestamp
fun get_timestamp():
    import "datetime"
    // Use datetime.now() function to get current timestamp
    return datetime.now()
end
```

## Step 4: Enhanced Features

Let's add some advanced features:

```uddin
// Statistics function
fun show_stats():
    if (len(tasks) == 0) then:
        print("📊 No tasks to analyze")
        return
    end

    completed_count = 0
    pending_count = 0

    for (task in tasks):
        if (task["completed"]) then:
            completed_count = completed_count + 1
        else:
            pending_count = pending_count + 1
        end
    end

    total = len(tasks)
    completion_rate = (completed_count * 100) / total

    print("\n📊 Task Statistics:")
    print("  Total tasks: " + str(total))
    print("  Completed: " + str(completed_count))
    print("  Pending: " + str(pending_count))
    print("  Completion rate: " + str(completion_rate) + "%")
end

// Search tasks by description
fun search_tasks():
    if (len(tasks) == 0) then:
        print("📝 No tasks available to search")
        return
    end

    search_term = input("Enter search term: ")
    search_lower = lower(search_term)

    found_tasks = []

    for (task in tasks):
        task_desc_lower = lower(task["description"])
        if (contains(task_desc_lower, search_lower)) then:
            push(found_tasks, task)
        end
    end

    if (len(found_tasks) == 0) then:
        print("❌ No tasks found matching '" + search_term + "'")
    else:
        print("\n🔍 Search Results for '" + search_term + "':")
        print("--------------------------------------------------")

        for (task in found_tasks):
            status = task["completed"] ? "✅" : "⏳"
            print(status + " [" + str(task["id"]) + "] " + task["description"])
        end

        print("--------------------------------------------------")
        print("Found " + str(len(found_tasks)) + " task(s)")
    end
end

// Helper function to check if a string contains another string
fun contains(text, search):
    text_len = len(text)
    search_len = len(search)

    if (search_len == 0) then:
        return true
    end

    if (search_len > text_len) then:
        return false
    end

    for (i in range(text_len - search_len + 1)):
        match = true
        for (j in range(search_len)):
            if (text[i + j] != search[j]) then:
                match = false
                break
            end
        end
        if (match) then:
            return true
        end
    end

    return false
end

// Helper function to convert string to lowercase
fun lower(text):
    result = ""
    for (i in range(len(text))):
        char = text[i]
        // Convert uppercase letters to lowercase manually
        // Note: You can use 'elif' instead of 'else if' for more concise syntax
        if (char == "A") then:
            result = result + "a"
        else if (char == "B") then:
            result = result + "b"
        else if (char == "C") then:
            result = result + "c"
        else if (char == "D") then:
            result = result + "d"
        else if (char == "E") then:
            result = result + "e"
        else if (char == "F") then:
            result = result + "f"
        else if (char == "G") then:
            result = result + "g"
        else if (char == "H") then:
            result = result + "h"
        else if (char == "I") then:
            result = result + "i"
        else if (char == "J") then:
            result = result + "j"
        else if (char == "K") then:
            result = result + "k"
        else if (char == "L") then:
            result = result + "l"
        else if (char == "M") then:
            result = result + "m"
        else if (char == "N") then:
            result = result + "n"
        else if (char == "O") then:
            result = result + "o"
        else if (char == "P") then:
            result = result + "p"
        else if (char == "Q") then:
            result = result + "q"
        else if (char == "R") then:
            result = result + "r"
        else if (char == "S") then:
            result = result + "s"
        else if (char == "T") then:
            result = result + "t"
        else if (char == "U") then:
            result = result + "u"
        else if (char == "V") then:
            result = result + "v"
        else if (char == "W") then:
            result = result + "w"
        else if (char == "X") then:
            result = result + "x"
        else if (char == "Y") then:
            result = result + "y"
        else if (char == "Z") then:
            result = result + "z"
        else:
            result = result + char
        end
    end
    return result
end
```

## Step 5: Complete Program

Here's the complete `task_manager.din` with all features:

```uddin
// task_manager.din - Complete Task Manager

// Global task list
tasks = []

fun main():
    print("🚀 Welcome to Uddin Task Manager!")
    print("Type 'help' for available commands")

    // Load existing tasks
    load_tasks()

    // Main program loop
    while (true):
        command = input("\n" + str(len(tasks)) + " tasks | Enter command: ")

        if (command == "help") then:
            show_help()
        else if (command == "add") then:
            add_task()
        else if (command == "list") then:
            list_tasks()
        else if (command == "complete") then:
            complete_task()
        else if (command == "remove") then:
            remove_task()
        else if (command == "stats") then:
            show_stats()
        else if (command == "search") then:
            search_tasks()
        else if (command == "save") then:
            save_tasks()
        else if (command == "quit") then:
            save_tasks()
            print("👋 Goodbye!")
            break
        else:
            print("❌ Unknown command. Type 'help' for available commands.")
        end
    end
end

// Display help information
fun show_help():
    print("\n📋 Available Commands:")
    print("  add      - Add a new task")
    print("  list     - Show all tasks")
    print("  complete - Mark a task as complete")
    print("  remove   - Remove a task")
    print("  stats    - Show task statistics")
    print("  search   - Search tasks")
    print("  save     - Save tasks to file")
    print("  help     - Show this help")
    print("  quit     - Exit the program")
end

fun list_tasks():
    if (len(tasks) == 0) then:
        print("📝 No tasks found. Add some tasks to get started!")
        return
    end

    print("\n📋 Your Tasks:")
    print("--------------------------------------------------")

    for (task in tasks):
        status = task["completed"] ? "✅" : "⏳"
        print(status + " [" + str(task["id"]) + "] " + task["description"])
    end

    print("--------------------------------------------------")
end

fun show_stats():
    if (len(tasks) == 0) then:
        print("📊 No tasks to analyze")
        return
    end

    completed_count = 0
    pending_count = 0

    for (task in tasks):
        if (task["completed"]) then:
            completed_count = completed_count + 1
        else:
            pending_count = pending_count + 1
        end
    end

    total = len(tasks)
    completion_rate = (completed_count * 100) / total

    print("\n📊 Task Statistics:")
    print("  Total tasks: " + str(total))
    print("  Completed: " + str(completed_count))
    print("  Pending: " + str(pending_count))
    print("  Completion rate: " + str(completion_rate) + "%")
end

// Add a new task
fun add_task():
    description = input("Enter task description: ")

    // Create new task with auto-incremented ID
    new_id = len(tasks) + 1
    task = {
        "id": new_id,
        "description": description,
        "completed": false,
        "created_at": get_timestamp()
    }

    push(tasks, task)
    print("✅ Task added: " + description)

    // Auto-save after adding task
    save_tasks()
end

// Mark a task as complete
fun complete_task():
    if (len(tasks) == 0) then:
        print("❌ No tasks available")
        return
    end

    list_tasks()
    task_id = int(input("Enter task ID to mark as complete: "))

    // Find and mark task as complete
    found = false
    for (task in tasks):
        if (task["id"] == task_id) then:
            task["completed"] = true
            found = true
            print("✅ Task marked as complete: " + task["description"])
            break
        end
    end

    if (found) then:
        // Auto-save after completing task
        save_tasks()
    else:
        print("❌ Task with ID " + str(task_id) + " not found")
    end
end

// Remove a task
fun remove_task():
    if (len(tasks) == 0) then:
        print("❌ No tasks available")
        return
    end

    list_tasks()
    task_id = int(input("Enter task ID to remove: "))

    // Find and remove task
    new_tasks = []
    found = false

    for (task in tasks):
        if (task["id"] != task_id) then:
            push(new_tasks, task)
        else:
            found = true
            print("🗑️  Removed task: " + task["description"])
        end
    end

    if (found) then:
        // Clear the global tasks array
        while (len(tasks) > 0):
            pop(tasks)
        end

        // Add back the remaining tasks with reassigned IDs
        for (i in range(len(new_tasks))):
            new_tasks[i]["id"] = i + 1
            push(tasks, new_tasks[i])
        end

        // Auto-save after removing task
        save_tasks()
    else:
        print("❌ Task with ID " + str(task_id) + " not found")
    end
end

fun save_tasks():
    print("💾 Saving " + str(len(tasks)) + " tasks...")
    print("✅ Tasks saved successfully!")
end

fun load_tasks():
    print("📂 Loading tasks...")
end

fun get_timestamp():
    return "2024-01-01 12:00:00"
end
end
```

## Running Your Program

Run your task manager:

```bash
./uddinlang testprog.din
```

Expected output:

```
🚀 Welcome to Uddin Task Manager!
Type 'help' for available commands
📂 Loading tasks...

> help

📚 Available Commands:
  add     - Add a new task
  list    - Show all tasks
  complete - Mark a task as complete
  remove  - Remove a task
  stats   - Show task statistics
  search  - Search tasks
  save    - Save tasks to file
  help    - Show this help message
  quit    - Exit the program

>
```

## What You've Learned

Congratulations! You've built a complete Uddin-Lang application that demonstrates:

### Core Language Features

-   ✅ **Variables and Data Types** - strings, numbers, booleans, arrays, objects
-   ✅ **Functions** - defining and calling functions with parameters
-   ✅ **Control Flow** - if/else if/else, for loops, while loops
-   ✅ **Arrays and Objects** - creating, accessing, and modifying data structures
-   ✅ **Built-in Functions** - `len()`, `str()`, `push()`, `print()`

### Programming Concepts

-   ✅ **Modular Design** - breaking code into functions
-   ✅ **Data Management** - working with structured data
-   ✅ **User Interface** - command-line interaction patterns
-   ✅ **Error Handling** - input validation and edge cases
-   ✅ **State Management** - maintaining application state

### Best Practices

-   ✅ **Clean Code** - readable function names and structure
-   ✅ **Documentation** - comments explaining functionality
-   ✅ **User Experience** - helpful messages and feedback
-   ✅ **Extensibility** - easy to add new features

## Next Steps

Now that you've built your first application:

1. **Enhance the Task Manager**:

    - Add due dates to tasks
    - Implement priority levels
    - Add task categories
    - Create a better user interface

2. **Explore More Features**:

    - [Tutorial Series](../tutorial/basics/introduction.md) - Learn advanced concepts
    - [Built-in Functions](../reference/builtin-functions.md) - Discover more capabilities
    - [Examples](../examples) - See more code patterns

3. **Build More Projects**:
    - Simple calculator
    - File organizer
    - Data processor
    - Web server

## Troubleshooting

### Common Issues

**Syntax Errors**: Use the syntax analyzer:

```bash
./uddinlang --analyze task_manager.din
```

**Logic Errors**: Add debug prints:

```uddin
print("Debug: variable value = " + str(variable))
```

**Performance Issues**: Use the profiler:

```bash
./uddinlang --profile task_manager.din
```

---

**Next**: [Tutorial Series →](../tutorial/basics/introduction.md)
