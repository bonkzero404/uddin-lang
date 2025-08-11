---
sidebar_position: 1
---

# Built-in Functions

Uddin-Lang provides a comprehensive set of built-in functions to help you build powerful applications efficiently.

## Type Conversion Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `str(value)` | Convert to string | `str(42)` → `"42"` | string |
| `int(value)` | Convert to integer | `int("42")` → `42` | int |
| `float(value)` | Convert to float | `float("3.14")` → `3.14` | float |
| `bool(value)` | Convert to boolean | `bool(1)` → `true` | bool |
| `char(value)` | Convert to character | `char(65)` → `"A"` | string |
| `rune(value)` | Convert to Unicode rune | `rune("A")` → `65` | int |
| `type(value)` | Get type name | `type(42)` → `"int"` | string |
| `typeof(value)` | Get detailed type info | `typeof([1,2,3])` → `"array[int]"` | string |

## String Functions

### Basic String Operations

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `len(string)` | String length | `len("hello")` → `5` | int |
| `substr(string, start, length)` | Extract substring | `substr("hello", 1, 3)` → `"ell"` | string |
| `split(string, delimiter)` | Split string | `split("a,b,c", ",")` → `["a","b","c"]` | array |
| `join(array, delimiter)` | Join array to string | `join(["a","b"], ",")` → `"a,b"` | string |
| `trim(string)` | Remove whitespace | `trim(" hello ")` → `"hello"` | string |
| `replace(string, old, new)` | Replace substring | `replace("hello", "l", "x")` → `"hexxo"` | string |

### String Case Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `upper(string)` | Convert to uppercase | `upper("hello")` → `"HELLO"` | string |
| `lower(string)` | Convert to lowercase | `lower("HELLO")` → `"hello"` | string |

### String Search Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `contains(string, substring)` | Check if contains | `contains("hello", "ell")` → `true` | bool |
| `starts_with(string, prefix)` | Check prefix | `starts_with("hello", "he")` → `true` | bool |
| `ends_with(string, suffix)` | Check suffix | `ends_with("hello", "lo")` → `true` | bool |

## Array Functions

### Core Array Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `len(array)` | Array length | `len([1,2,3])` → `3` | int |
| `append(array, items...)` | Add items | `append([1,2], 3, 4)` → `[1,2,3,4]` | array |
| `slice(array, start, end)` | Extract slice | `slice([1,2,3,4], 1, 3)` → `[2,3]` | array |
| `sort(array)` | Sort in place | `sort([3,1,2])` → `[1,2,3]` | array |
| `range(n)` or `range(start, stop)` | Create range | `range(3)` → `[0,1,2]`<br />`range(1, 4)` → `[1,2,3]` | array |
| `find(array, value)` | Find index | `find([1,2,3], 2)` → `1` | int |
| `contains(array, value)` | Check membership | `contains([1,2,3], 2)` → `true` | bool |

### Functional Programming Methods

| Function | Description | Example |
|----------|-------------|----------|
| `map(array, function)` | Transform each element | `map([1,2,3], fun(x): return x*2 end)` → `[2,4,6]` |
| `filter(array, function)` | Filter elements by condition | `filter([1,2,3,4], fun(x): return x%2==0 end)` → `[2,4]` |
| `reduce(array, function, init)` | Reduce to single value | `reduce([1,2,3], fun(a,x): return a+x end, 0)` → `6` |
| `reverse(array)` | Reverse array in-place | `reverse([1,2,3])` modifies to `[3,2,1]` |
| `push(array, element)` | Add element to end | `push([1,2], 3)` modifies to `[1,2,3]` |
| `pop(array)` | Remove and return last element | `pop([1,2,3])` → `3`, array becomes `[1,2]` |
| `shift(array)` | Remove and return first element | `shift([1,2,3])` → `1`, array becomes `[2,3]` |
| `unshift(array, element)` | Add element to beginning | `unshift([2,3], 1)` modifies to `[1,2,3]` |
| `index_of(array, element)` | Find first index of element | `index_of([1,2,3,2], 2)` → `1` |
| `last_index_of(array, element)` | Find last index of element | `last_index_of([1,2,3,2], 2)` → `3` |

## Data Structures

### Set (Unique Collection)

| Function | Description | Example |
|----------|-------------|----------|
| `set_new()` | Create new empty set | `my_set = set_new()` |
| `set_add(set, elem)` | Add element (ignores duplicates) | `set_add(my_set, 1)` |
| `set_has(set, elem)` | Check if element exists | `set_has(my_set, 1)` → `true` |
| `set_remove(set, elem)` | Remove element | `set_remove(my_set, 1)` → `true` |
| `set_size(set)` | Get number of elements | `set_size(my_set)` → `3` |
| `set_to_array(set)` | Convert set to array | `set_to_array(my_set)` → `[1,2,3]` |

### Stack (LIFO - Last In, First Out)

| Function | Description | Example |
|----------|-------------|----------|
| `stack_new()` | Create new empty stack | `my_stack = stack_new()` |
| `stack_push(stack, elem)` | Add element to top | `stack_push(my_stack, "item")` |
| `stack_pop(stack)` | Remove and return top element | `stack_pop(my_stack)` → `"item"` |
| `stack_peek(stack)` | View top element without removing | `stack_peek(my_stack)` → `"item"` |
| `stack_size(stack)` | Get number of elements | `stack_size(my_stack)` → `3` |
| `stack_empty(stack)` | Check if stack is empty | `stack_empty(my_stack)` → `false` |

### Queue (FIFO - First In, First Out)

| Function | Description | Example |
|----------|-------------|----------|
| `queue_new()` | Create new empty queue | `my_queue = queue_new()` |
| `queue_enqueue(queue, elem)` | Add element to back | `queue_enqueue(my_queue, "task")` |
| `queue_dequeue(queue)` | Remove and return front element | `queue_dequeue(my_queue)` → `"task"` |
| `queue_front(queue)` | View front element without removing | `queue_front(my_queue)` → `"task"` |
| `queue_size(queue)` | Get number of elements | `queue_size(my_queue)` → `3` |
| `queue_empty(queue)` | Check if queue is empty | `queue_empty(my_queue)` → `false` |

## Math Functions

### Basic Math Operations

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `abs(x)` | Absolute value | `abs(-5)` → `5` | int/float |
| `max(a, b, ...)` | Maximum value | `max(1, 5, 3)` → `5` | int/float |
| `min(a, b, ...)` | Minimum value | `min(1, 5, 3)` → `1` | int/float |
| `pow(base, exp)` | Power function | `pow(2, 3)` → `8` | int/float |
| `sqrt(x)` | Square root | `sqrt(16)` → `4.0` | float |
| `cbrt(x)` | Cube root | `cbrt(27)` → `3.0` | float |

### Rounding Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `round(x)` | Round to nearest integer | `round(3.7)` → `4` | int |
| `round(x, n)` | Round to n decimal places | `round(3.14159, 2)` → `3.14` | float |
| `floor(x)` | Round down (floor) | `floor(3.7)` → `3` | int |
| `ceil(x)` | Round up (ceiling) | `ceil(3.2)` → `4` | int |
| `trunc(x)` | Truncate decimal part | `trunc(3.7)` → `3` | int |

### Trigonometric Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `sin(x)` | Sine (radians) | `sin(PI/2)` → `1.0` | float |
| `cos(x)` | Cosine (radians) | `cos(0)` → `1.0` | float |
| `tan(x)` | Tangent (radians) | `tan(PI/4)` → `1.0` | float |
| `asin(x)` | Arc sine (returns radians) | `asin(1)` → `1.5708` | float |
| `acos(x)` | Arc cosine (returns radians) | `acos(1)` → `0.0` | float |
| `atan(x)` | Arc tangent (returns radians) | `atan(1)` → `0.7854` | float |
| `atan2(y, x)` | Arc tangent of y/x | `atan2(1, 1)` → `0.7854` | float |

### Hyperbolic Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `sinh(x)` | Hyperbolic sine | `sinh(1.0)` → `1.175` | float |
| `cosh(x)` | Hyperbolic cosine | `cosh(1.0)` → `1.543` | float |
| `tanh(x)` | Hyperbolic tangent | `tanh(1.0)` → `0.762` | float |

### Logarithmic Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `log(x)` | Natural logarithm (ln) | `log(E)` → `1.0` | float |
| `log10(x)` | Base-10 logarithm | `log10(100)` → `2.0` | float |
| `log2(x)` | Base-2 logarithm | `log2(8)` → `3.0` | float |
| `logb(x, base)` | Logarithm with custom base | `logb(125, 5)` → `3.0` | float |
| `exp(x)` | Exponential function (e^x) | `exp(1)` → `2.718...` | float |
| `exp2(x)` | Base-2 exponential (2^x) | `exp2(3)` → `8.0` | float |

### Statistical Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `sum(array)` | Sum of array elements | `sum([1, 2, 3, 4])` → `10` | int/float |
| `mean(array)` | Arithmetic mean (average) | `mean([1, 2, 3, 4])` → `2.5` | float |
| `median(array)` | Median value | `median([1, 2, 3, 4, 5])` → `3.0` | float |
| `mode(array)` | Most frequent value | `mode([1, 2, 2, 3])` → `2` | any |
| `std_dev(array)` | Standard deviation | `std_dev([1, 2, 3, 4])` → `1.29` | float |
| `variance(array)` | Variance | `variance([1, 2, 3, 4])` → `1.67` | float |

### Number Theory Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `gcd(a, b)` | Greatest common divisor | `gcd(12, 8)` → `4` | int |
| `lcm(a, b)` | Least common multiple | `lcm(12, 8)` → `24` | int |
| `factorial(n)` | Factorial (n!) | `factorial(5)` → `120` | int |
| `fibonacci(n)` | Nth Fibonacci number | `fibonacci(7)` → `13` | int |
| `is_prime(n)` | Check if number is prime | `is_prime(17)` → `true` | bool |
| `prime_factors(n)` | List of prime factors | `prime_factors(12)` → `[2,2,3]` | array |

### Random Number Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `random()` | Random float between 0 and 1 | `random()` → `0.7234` | float |
| `random_int(min, max)` | Random integer in range | `random_int(1, 10)` → `7` | int |
| `random_float(min, max)` | Random float in range | `random_float(1.0, 2.0)` → `1.45` | float |
| `random_choice(array)` | Random element from array | `random_choice([1,2,3])` → `2` | any |
| `shuffle(array)` | Shuffle array in place | `shuffle([1,2,3])` → `[3,1,2]` | array |
| `seed_random(seed)` | Set random seed | `seed_random(42)` | null |

### Mathematical Constants

| Constant | Description | Value |
|----------|-------------|-------|
| `PI` | Pi (π) | `3.14159265359` |
| `E` | Euler's number (e) | `2.71828182846` |
| `TAU` | Tau (2π) | `6.28318530718` |
| `PHI` | Golden ratio (φ) | `1.61803398875` |
| `LN2` | Natural logarithm of 2 | `0.69314718056` |
| `LN10` | Natural logarithm of 10 | `2.30258509299` |
| `SQRT2` | Square root of 2 | `1.41421356237` |
| `SQRT3` | Square root of 3 | `1.73205080757` |

## I/O Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `print(values...)` | Print to console | `print("Hello", "World")` | void |
| `input(prompt)` | Read user input | `name = input("Enter name: ")` | string |
| `read_file(path)` | Read file content | `content = read_file("data.txt")` | string |
| `write_file(path, content)` | Write to file | `write_file("out.txt", "Hello")` | bool |

## File System Operations

### File Operations

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `read_file(path)` | Read file content | `content = read_file("data.txt")` | string |
| `write_file(path, content)` | Write content to file | `write_file("output.txt", "Hello World")` | bool |
| `file_exists(path)` | Check if file/directory exists | `file_exists("config.json")` → `true` | bool |
| `file_size(path)` | Get file size in bytes | `file_size("data.txt")` → `1024` | int |
| `file_modified(path)` | Get last modification time | `file_modified("log.txt")` → `"2024-01-15"` | string |
| `file_permissions(path)` | Get file permissions | `file_permissions("script.sh")` → `"755"` | string |
| `copy_file(source, destination)` | Copy file to new location | `copy_file("src.txt", "backup.txt")` | bool |
| `move_file(source, destination)` | Move/rename file | `move_file("old.txt", "new.txt")` | bool |
| `delete_file(path)` | Delete file | `delete_file("temp.txt")` | bool |

### Directory Operations

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `mkdir(path)` | Create directory | `mkdir("logs")` | bool |
| `rmdir(path)` | Remove empty directory | `rmdir("temp")` | bool |
| `list_dir(path)` | List directory contents | `list_dir(".")` → `["file1", "dir1"]` | array |
| `getcwd()` | Get current working directory | `getcwd()` → `"/home/user/project"` | string |
| `chdir(path)` | Change working directory | `chdir("/tmp")` | bool |

### Path Operations

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `path_join(parts...)` | Join path components | `path_join("home", "user", "file.txt")` | string |
| `path_dirname(path)` | Get directory name | `path_dirname("/home/user/file.txt")` | string |
| `path_basename(path)` | Get base filename | `path_basename("/home/user/file.txt")` | string |
| `path_ext(path)` | Get file extension | `path_ext("document.pdf")` → `".pdf"` | string |

## JSON Functions

| Function | Description | Example |
|----------|-------------|----------|
| `json_parse(json_string)` | Parse JSON string to Uddin-Lang value | `data = json_parse('{"name": "John", "age": 30}')` |
| `json_stringify(value)` | Convert Uddin-Lang value to JSON string | `json_str = json_stringify({name: "Alice", age: 25})` |

### JSON Type Mapping

| JSON Type | Uddin-Lang Type | Example |
|-----------|-----------------|----------|
| `object` | `map[string]Value` | `{"key": "value"}` → `{key: "value"}` |
| `array` | `[]Value` | `[1, 2, 3]` → `[1, 2, 3]` |
| `string` | `string` | `"hello"` → `"hello"` |
| `number` | `int` or `float64` | `42` → `42`, `3.14` → `3.14` |
| `boolean` | `bool` | `true` → `true` |
| `null` | `null` | `null` → `null` |

## XML Processing

| Function | Description | Example |
|----------|-------------|----------|
| `xml_parse(xml_string)` | Parse XML string to Uddin-Lang value | `data = xml_parse('<person><name>John</name></person>')` |
| `xml_stringify(value)` | Convert Uddin-Lang value to XML string | `xml_str = xml_stringify({person: {name: "Alice", age: "25"}})` |

### XML Structure Mapping

| XML Feature | Uddin-Lang Representation | Example |
|-------------|---------------------------|----------|
| Root Element | Map key | `<root>...</root>` → `{"root": {...}}` |
| Child Elements | Map properties | `<name>John</name>` → `{"name": "John"}` |
| Attributes | `@attributes` object | `<item id="1">` → `{"@attributes": {"id": "1"}}` |
| Text Content | String value | `<title>Book</title>` → `{"title": "Book"}` |
| Multiple Elements | Array | `<item>1</item><item>2</item>` → `[1, 2]` |

## Networking Functions

### HTTP Client Functions

| Function | Description | Example |
|----------|-------------|----------|
| `http_get(url)` | HTTP GET request | `http_get("https://api.example.com/data")` |
| `http_post(url, data)` | HTTP POST request | `http_post("https://api.example.com", data)` |
| `http_put(url, data)` | HTTP PUT request | `http_put("https://api.example.com/1", data)` |
| `http_delete(url)` | HTTP DELETE request | `http_delete("https://api.example.com/1")` |
| `http_request(method, url, data)` | Generic HTTP request | `http_request("PATCH", url, data)` |

### HTTP Server Functions

| Function | Description | Example |
|----------|-------------|----------|
| `http_server_start(port, server_id?)` | Start HTTP server on specified port | `server = http_server_start(8080, "main")` |
| `http_server_stop(server_id?)` | Stop HTTP server | `http_server_stop("main")` |
| `http_server_route(method, path, handler, server_id?)` | Register route handler | `http_server_route("GET", "/api", my_handler, "main")` |
| `http_response(res, status, headers?, body?)` | Send HTTP response | `http_response(res, 200, {"Content-Type": "text/plain"}, "Hello")` |

### Network Utilities

| Function | Description | Example |
|----------|-------------|----------|
| `net_resolve(hostname)` | Resolve hostname to IP addresses | `net_resolve("google.com")` → `["142.250.191.14"]` |
| `net_ping(host, port, timeout)` | Test connectivity to host:port | `net_ping("google.com", 80, 3000)` → `{"success": true, "time": 45}` |

### TCP Functions

| Function | Description | Example |
|----------|-------------|----------|
| `tcp_connect(host, port)` | Create TCP client connection | `conn = tcp_connect("localhost", 8080)` |
| `tcp_listen(port)` | Create TCP server listener | `listener = tcp_listen(8080)` |
| `tcp_accept(listener)` | Accept incoming TCP connection | `client = tcp_accept(listener)` |
| `tcp_read(connection)` | Read data from TCP connection | `data = tcp_read(conn)` |
| `tcp_write(connection, data)` | Write data to TCP connection | `tcp_write(conn, "Hello Server!")` |
| `tcp_close(connection)` | Close TCP connection/listener | `tcp_close(conn)` |

### UDP Functions

| Function | Description | Example |
|----------|-------------|----------|
| `udp_connect(host, port)` | Create UDP client connection | `conn = udp_connect("localhost", 8080)` |
| `udp_listen(port)` | Create UDP server listener | `listener = udp_listen(8080)` |
| `udp_read(connection)` | Read data from UDP connection | `data = udp_read(conn)` |
| `udp_write(connection, data)` | Write data to UDP connection | `udp_write(conn, "Hello Server!")` |
| `udp_close(connection)` | Close UDP connection/listener | `udp_close(conn)` |

## Utility Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `sign(x)` | Sign of number (-1, 0, 1) | `sign(-5)` → `-1` | int |
| `clamp(x, min, max)` | Clamp value to range | `clamp(15, 1, 10)` → `10` | int/float |
| `lerp(a, b, t)` | Linear interpolation | `lerp(0, 10, 0.5)` → `5.0` | float |
| `degrees(radians)` | Convert radians to degrees | `degrees(PI)` → `180.0` | float |
| `radians(degrees)` | Convert degrees to radians | `radians(180)` → `3.14159` | float |
| `is_nan(x)` | Check if value is NaN | `is_nan(0.0/0.0)` → `true` | bool |
| `is_infinite(x)` | Check if value is infinite | `is_infinite(1.0/0.0)` → `true` | bool |
| `date_now()` | Current timestamp | `date_now()` → `"2025-06-26T14:30:00Z"` | string |
| `time_now()` | Current Unix timestamp in milliseconds | `time_now()` → `1640995445123` | int |
| `date_format(date, fmt)` | Format date | `date_format(date_now(), "YYYY-MM-DD")` | string |
| `date_parse(date_str, layout)` | Parse date string | `date_parse("2023-01-01", "2006-01-02")` | int |
| `date_format_new(timestamp, layout)` | Format timestamp with layout | `date_format_new(time_now(), "2006-01-02 15:04:05")` | string |
| `date_add(timestamp, duration)` | Add duration to timestamp | `date_add(time_now(), "24h")` | int |
| `date_subtract(timestamp, duration)` | Subtract duration from timestamp | `date_subtract(time_now(), "1h30m")` | int |
| `date_diff(timestamp1, timestamp2)` | Calculate difference between timestamps | `date_diff(time2, time1)` | int |
| `date_between(timestamp, start, end)` | Check if timestamp is between two dates | `date_between(now, start, end)` | bool |
| `date_compare(timestamp1, timestamp2)` | Compare two timestamps | `date_compare(time1, time2)` | int |

## Regular Expression Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `is_regex_match(pattern, text)` | Check if text matches regex pattern | `is_regex_match("^[0-9]+$", "123")` → `true` | bool |
| `regex_match(text, pattern)` | Check if text matches regex pattern | `regex_match("hello@example.com", email_pattern)` | bool |
| `regex_find(text, pattern)` | Find first match of regex pattern | `regex_find("Phone: 123-456-7890", "[0-9-]+")` → `"123-456-7890"` | string |
| `regex_find_all(text, pattern)` | Find all matches of regex pattern | `regex_find_all(text, email_pattern)` | array |
| `regex_replace(text, pattern, replacement)` | Replace regex matches | `regex_replace("hello world", "world", "universe")` | string |
| `regex_split(text, pattern)` | Split text by regex pattern | `regex_split("a,b;c", "[,;]")` → `["a", "b", "c"]` | array |

## System Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `exit(code)` | Exit program with code | `exit(0)` | void |

## Rule Engine - Fact Database Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `fact_assert(type, id, data)` | Add fact to database | `fact_assert("person", "john", {"age": 30})` | bool |
| `fact_retract(type, id, data)` | Remove fact from database | `fact_retract("person", "john", {})` | bool |
| `fact_query(type, id, pattern)` | Query facts with pattern | `fact_query("person", null, {"city": "Jakarta"})` | array |
| `fact_exists(type, id, pattern)` | Check if fact exists | `fact_exists("person", "john", {})` | bool |
| `fact_count(type, id, pattern)` | Count matching facts | `fact_count("person", null, {})` | int |
| `fact_clear()` | Clear all facts | `fact_clear()` | void |
| `fact_get_all()` | Get all facts | `fact_get_all()` | array |

## Complex Event Processing Functions

| Function | Description | Example | Return Type |
|----------|-------------|---------|-------------|
| `event_emit(type, data)` | Emit event | `event_emit("user_login", {"user": "john"})` | void |
| `event_define_pattern(name, pattern)` | Define event pattern | `event_define_pattern("login_pattern", pattern)` | bool |
| `event_get_window(name)` | Get event window | `event_get_window("recent_events")` | array |
| `event_clear()` | Clear all events | `event_clear()` | void |
| `event_count(type)` | Count events by type | `event_count("user_login")` | int |

For more detailed examples and advanced usage, see the [Tutorial](../tutorial/basics/introduction.md) section.
