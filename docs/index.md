---
layout: default
title: Uddin Programming Language
---

# Uddin Programming Language

Welcome to the official documentation of Uddin Programming Language - a modern programming language designed for ease of use and flexibility.

## About Uddin Lang

Uddin Lang is an interpretive programming language that combines easy-to-understand syntax with advanced features for modern application development. The language supports various programming paradigms and comes with a rich set of built-in libraries.

## Key Features

- **Simple Syntax**: Easy to learn and understand
- **Comprehensive Built-in Functions**: Mathematics, strings, arrays, networking, and file system
- **Modern Data Structures**: Set, Stack, Queue, and Array with advanced methods
- **Networking Support**: HTTP client/server, TCP, UDP
- **File System Operations**: Comprehensive file and directory operations
- **JSON & XML Processing**: Parsing and manipulation of structured data
- **Rule Engine**: Rule system for complex business logic

## Quick Start

### Installation

```bash
# Clone repository
git clone https://github.com/bonkzero404/uddin-lang.git
cd uddin-lang

# Build interpreter
go build -o uddin main.go

# Run your first program
./uddin examples/01_hello_world.din
```

### Your First Program

```uddin
// Hello World in Uddin Lang
print("Hello, World!")

// Variables and operations
name = "Uddin"
age = 25
print("Name: " + name + ", Age: " + str(age))

// Functions
func greet(name) {
    return "Hello, " + name + "!"
}

print(greet("Programmer"))
```

## Documentation Navigation

### 📚 [Complete Tutorial](tutorial/)
Step-by-step guide to learn Uddin Lang from basics to advanced.

### 📖 [Language Reference](reference/)
Complete documentation of syntax, operators, and language structures.

### 🔧 [Built-in Functions](builtin-functions/)
Complete list of all available built-in functions.

### 💡 [Program Examples](examples/)
Collection of example programs for various use cases.

### 🚀 [Advanced Guide](advanced/)
Advanced topics such as performance optimization and best practices.

## Featured Examples

### File System Operations
```uddin
// Create and write file
write_file("data.txt", "Hello from Uddin!")

// Read file
content = read_file("data.txt")
print(content)

// Directory operations
make_dir("my_folder")
files = list_dir(".")
print("Files:", files)
```

### HTTP Client
```uddin
// GET request
response = http_get("https://api.github.com/users/octocat")
data = json_parse(response)
print("User:", data["name"])

// POST request
payload = json_stringify({"name": "test", "value": 123})
response = http_post("https://httpbin.org/post", payload, {"Content-Type": "application/json"})
```

### Modern Data Structures
```uddin
// Set operations
my_set = set_create()
set_add(my_set, "apple")
set_add(my_set, "banana")
print("Set contains apple:", set_contains(my_set, "apple"))

// Stack operations
my_stack = stack_create()
stack_push(my_stack, "first")
stack_push(my_stack, "second")
item = stack_pop(my_stack)
print("Popped:", item)
```

## Community and Contribution

- **GitHub**: [Uddin Lang Repository](https://github.com/bonkzero404/uddin-lang)
- **Issues**: Report bugs or request features
- **Discussions**: Discuss with the community
- **Contributing**: Contribution guide for developers

## License

Uddin Programming Language is released under the MIT license. See the [LICENSE](https://github.com/bonkzero404/uddin-lang/blob/main/LICENSE) file for complete details.

---

*This documentation is continuously updated. For the latest version, visit the [GitHub repository](https://github.com/bonkzero404/uddin-lang).*