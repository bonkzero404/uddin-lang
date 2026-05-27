---
id: modules
title: Module System
sidebar_label: Modules
---

# Module System

Uddin-Lang provides a **stdlib module system** that organises domain-specific functionality into opt-in namespaced modules. Instead of polluting the global scope with hundreds of built-in names, you import only the modules you need.

## Import Syntax

```din
import "module"          // binds as module.function()
import "module" as alias // binds as alias.function()
```

**Examples:**

```din
import "database"
database.connect("postgres://...")

import "database" as db
db.connect("postgres://...")
```

The default variable name is the module name. Use `as` to shorten it.

---

## Available Modules

| Module | Import | Purpose |
|--------|--------|---------|
| `regex` | `import "regex"` | Regular expression matching and manipulation |
| `datetime` | `import "datetime"` | Date/time operations and formatting |
| `json` | `import "json"` | JSON parsing and serialization |
| `fs` | `import "fs"` | File system operations |
| `http` | `import "http"` | HTTP client/server and TCP/UDP networking |
| `waf` | `import "waf"` | Web Application Firewall request inspection |
| `cdc` | `import "cdc"` | Change Data Capture / event streaming |
| `fact` | `import "fact"` | In-memory fact/knowledge database |
| `database` | `import "database"` | SQL database connectivity |

---

## `regex`

```din
import "regex"
```

| Function | Signature | Description |
|----------|-----------|-------------|
| `regex.is_match` | `(pattern, text)` | Returns `true` if `text` matches `pattern` |
| `regex.match` | `(text, pattern)` | Returns `true` if `text` matches `pattern` |
| `regex.find` | `(text, pattern)` | Returns first match string or `null` |
| `regex.find_all` | `(text, pattern [, limit])` | Returns array of all matches |
| `regex.replace` | `(text, pattern, replacement)` | Returns string with matches replaced |
| `regex.split` | `(text, pattern [, limit])` | Splits string by pattern |

**Example:**

```din
import "regex"

if regex.is_match("^\d+$", "42") then
    print("numeric")
end

matches = regex.find_all("cat 123 dog 456", "\d+")
print(matches)  // ["123", "456"]

clean = regex.replace("hello   world", "\s+", " ")
print(clean)    // "hello world"
```

---

## `datetime`

```din
import "datetime"
```

| Function | Signature | Description |
|----------|-----------|-------------|
| `datetime.now` | `()` | Current time in RFC3339 format |
| `datetime.time_now` | `()` | Current Unix timestamp (milliseconds) |
| `datetime.sleep` | `(ms)` | Sleep for `ms` milliseconds |
| `datetime.format` | `(date, layout)` | Format a date string |
| `datetime.parse` | `(date, layout)` | Parse a date string |
| `datetime.format_enhanced` | `(date, layout, locale?)` | Format with locale support |
| `datetime.add` | `(date, amount, unit)` | Add duration to a date |
| `datetime.subtract` | `(date, amount, unit)` | Subtract duration from a date |
| `datetime.diff` | `(date1, date2, unit)` | Difference between two dates |
| `datetime.between` | `(date, start, end)` | Check if date is in range |
| `datetime.compare` | `(date1, date2)` | Compare two dates: -1, 0, or 1 |

**Example:**

```din
import "datetime"

now = datetime.now()
print(now)                                    // "2026-05-28T10:00:00Z"
print(datetime.add(now, 7, "days"))           // 7 days later
print(datetime.diff(now, "2026-01-01", "days")) // days since Jan 1
```

---

## `json`

```din
import "json"
```

| Function | Signature | Description |
|----------|-----------|-------------|
| `json.parse` | `(string)` | Parse JSON string → value |
| `json.stringify` | `(value)` | Serialize value → JSON string |

**Example:**

```din
import "json"

data = json.parse('{"name": "Alice", "age": 30}')
print(data["name"])          // Alice

out = json.stringify({"x": 1, "y": [2, 3]})
print(out)                   // {"x":1,"y":[2,3]}
```

---

## `fs`

```din
import "fs"
```

| Function | Signature | Description |
|----------|-----------|-------------|
| `fs.read` | `(path)` | Read file contents as string |
| `fs.write` | `(path, content)` | Write string to file |
| `fs.exists` | `(path)` | Check if file/directory exists |
| `fs.size` | `(path)` | File size in bytes |
| `fs.modified` | `(path)` | Last modified time (RFC3339) |
| `fs.permissions` | `(path)` | File permissions as octal string |
| `fs.mkdir` | `(path [, recursive])` | Create directory |
| `fs.rmdir` | `(path)` | Remove directory |
| `fs.list_dir` | `(path)` | List directory entries |
| `fs.copy` | `(src, dst)` | Copy file |
| `fs.move` | `(src, dst)` | Move/rename file |
| `fs.delete` | `(path)` | Delete file |
| `fs.path_join` | `(parts...)` | Join path components |
| `fs.path_dirname` | `(path)` | Directory component of path |
| `fs.path_basename` | `(path)` | Base name component of path |
| `fs.path_ext` | `(path)` | File extension |
| `fs.getcwd` | `()` | Current working directory |
| `fs.chdir` | `(path)` | Change working directory |

**Example:**

```din
import "fs"

if fs.exists("config.json") then
    content = fs.read("config.json")
    print(content)
end

fs.write("output.txt", "Hello, world!")
files = fs.list_dir(".")
print(files)
```

---

## `http`

```din
import "http"
```

### HTTP Client

| Function | Signature | Description |
|----------|-----------|-------------|
| `http.get` | `(url [, headers])` | HTTP GET |
| `http.post` | `(url, data [, headers])` | HTTP POST |
| `http.put` | `(url, data [, headers])` | HTTP PUT |
| `http.delete` | `(url [, headers])` | HTTP DELETE |
| `http.request` | `(method, url, data, headers)` | Generic HTTP request |

### HTTP Server

| Function | Signature | Description |
|----------|-----------|-------------|
| `http.server_start` | `(id, port)` | Start HTTP server |
| `http.server_stop` | `(id)` | Stop HTTP server |
| `http.server_route` | `(id, method, path, handler)` | Register route handler |
| `http.response` | `(status, body [, headers])` | Build response object |

### TCP/UDP

| Function | Signature | Description |
|----------|-----------|-------------|
| `http.tcp_connect` | `(host, port)` | TCP client connection |
| `http.tcp_listen` | `(host, port)` | TCP server socket |
| `http.tcp_accept` | `(socket)` | Accept TCP connection |
| `http.tcp_read` | `(conn [, size])` | Read from TCP connection |
| `http.tcp_write` | `(conn, data)` | Write to TCP connection |
| `http.tcp_close` | `(conn)` | Close TCP connection |
| `http.udp_connect` | `(host, port)` | UDP connection |
| `http.udp_listen` | `(host, port)` | UDP server socket |
| `http.udp_read` | `(conn [, size])` | Read from UDP |
| `http.udp_write` | `(conn, data)` | Write to UDP |
| `http.udp_close` | `(conn)` | Close UDP |
| `http.net_resolve` | `(host)` | DNS resolve |
| `http.net_ping` | `(host)` | ICMP ping |

**Example:**

```din
import "http"

resp = http.get("https://api.example.com/data")
print(resp["status"])   // 200
print(resp["body"])
```

---

## `waf`

```din
import "waf"
```

Most WAF functions read from the `_waf_ctx` map injected by the host (except `path_match` and `cidr_match` which take explicit arguments). See [WAFio integration](./library#wafio) for how to populate it.

| Function | Signature | Description |
|----------|-----------|-------------|
| `waf.header` | `(name)` | Request header value (case-insensitive) or `null` |
| `waf.query` | `(name)` | URL query parameter value or `null` |
| `waf.body_contains` | `(needle)` | `true` if request body contains substring |
| `waf.cidr_match` | `(ip, cidr)` | `true` if IP is in CIDR range |
| `waf.path_match` | `(pattern, path)` | Glob-match path against pattern |
| `waf.score` | `()` | Current threat score (float) |
| `waf.action` | `()` | Current action string (`"ALLOW"`, `"LOG"`, `"BLOCK"`) |
| `waf.detected` | `(category)` | `true` if category was detected |
| `waf.detected_any` | `()` | `true` if any categories were detected |
| `waf.detected_list` | `()` | Array of detected category strings |
| `waf.return` | `(action)` | Exit rule with action (`"ALLOW"`, `"LOG"`, `"BLOCK"`) |

**Example:**

```din
import "waf"

if waf.detected_any() then
    if waf.cidr_match(waf.header("X-Real-IP"), "10.0.0.0/8") then
        waf.return("ALLOW")
    end
    waf.return("BLOCK")
end
waf.return("ALLOW")
```

---

## `cdc`

```din
import "cdc"
```

Change Data Capture / Complex Event Processing — an in-process event stream.

| Function | Signature | Description |
|----------|-----------|-------------|
| `cdc.emit` | `(type [, data])` | Emit event of given type |
| `cdc.define_pattern` | `(name, sequence [, window])` | Define event pattern |
| `cdc.get_window` | `(duration [, type])` | Events within time window |
| `cdc.clear` | `([type])` | Clear all events or events of given type |
| `cdc.count` | `([type])` | Count all events or events of given type |

**Example:**

```din
import "cdc"

cdc.emit("user_login", {"user": "alice"})
cdc.emit("page_view", {"path": "/dashboard"})

print(cdc.count())           // 2
print(cdc.count("user_login")) // 1

recent = cdc.get_window("5m")
print(recent)
```

---

## `fact`

```din
import "fact"
```

In-memory knowledge base keyed by `category:key`.

| Function | Signature | Description |
|----------|-----------|-------------|
| `fact.assert` | `(category, key [, value])` | Store a fact (default value: `true`) |
| `fact.retract` | `(category [, key])` | Remove fact or entire category |
| `fact.query` | `(category [, key])` | Retrieve fact or all facts in category |
| `fact.exists` | `(category, key)` | Check if a fact exists |
| `fact.count` | `([category])` | Count facts in category or total |
| `fact.clear` | `([category])` | Clear facts in category or all facts |
| `fact.get_all` | `()` | Return all facts |

**Example:**

```din
import "fact"

fact.assert("user", "alice", {"role": "admin", "active": true})
fact.assert("user", "bob")

if fact.exists("user", "alice") then
    alice = fact.query("user", "alice")
    print(alice["role"])   // admin
end

print(fact.count("user")) // 2
```

---

## `database`

```din
import "database"
```

SQL database connectivity supporting PostgreSQL, MySQL, SQLite, and MSSQL.

| Function | Signature | Description |
|----------|-----------|-------------|
| `database.connect` | `(dsn)` | Open connection, returns connection ID |
| `database.query` | `(conn, sql [, args...])` | Execute SELECT, returns array of rows |
| `database.exec` | `(conn, sql [, args...])` | Execute INSERT/UPDATE/DELETE |
| `database.close` | `(conn)` | Close connection |
| `database.begin` | `(conn)` | Begin transaction |
| `database.commit` | `(tx)` | Commit transaction |
| `database.rollback` | `(tx)` | Rollback transaction |
| `database.prepare` | `(conn, sql)` | Prepare statement |
| `database.ping` | `(conn)` | Ping connection |
| `database.stats` | `(conn)` | Connection pool statistics |

**DSN formats:**

```
postgres://user:pass@host:5432/dbname
mysql://user:pass@tcp(host:3306)/dbname
sqlite:///path/to/file.db
sqlserver://user:pass@host:1433?database=dbname
```

**Example:**

```din
import "database" as db

conn = db.connect("postgres://user:pass@localhost/mydb")

rows = db.query(conn, "SELECT id, name FROM users WHERE active = $1", true)
for row in rows then
    print(row["name"])
end

db.exec(conn, "UPDATE users SET last_seen = NOW() WHERE id = $1", 42)
db.close(conn)
```

### Transactions

```din
import "database" as db

conn = db.connect("postgres://...")
tx = db.begin(conn)

try
    db.exec(tx, "INSERT INTO orders VALUES ($1, $2)", 1, "product-a")
    db.exec(tx, "UPDATE inventory SET qty = qty - 1 WHERE id = $1", 1)
    db.commit(tx)
catch(err)
    db.rollback(tx)
    print("Transaction failed:", err)
end
```

---

## Migration from Flat Globals

If you were using the old flat built-in names (pre-v1.2.0), here is the mapping:

| Old (removed) | New |
|---------------|-----|
| `is_regex_match(p, s)` | `regex.is_match(p, s)` |
| `regex_match(t, p)` | `regex.match(t, p)` |
| `regex_find(t, p)` | `regex.find(t, p)` |
| `regex_find_all(t, p)` | `regex.find_all(t, p)` |
| `regex_replace(t, p, r)` | `regex.replace(t, p, r)` |
| `regex_split(t, p)` | `regex.split(t, p)` |
| `date_now()` | `datetime.now()` |
| `time_now()` | `datetime.time_now()` |
| `sleep(ms)` | `datetime.sleep(ms)` |
| `date_format(d, l)` | `datetime.format(d, l)` |
| `json_parse(s)` | `json.parse(s)` |
| `json_stringify(v)` | `json.stringify(v)` |
| `read_file(p)` | `fs.read(p)` |
| `write_file(p, c)` | `fs.write(p, c)` |
| `file_exists(p)` | `fs.exists(p)` |
| `http_get(url)` | `http.get(url)` |
| `http_post(url, d)` | `http.post(url, d)` |
| `tcp_connect(h, p)` | `http.tcp_connect(h, p)` |
| `event_emit(t, d)` | `cdc.emit(t, d)` |
| `event_count(t)` | `cdc.count(t)` |
| `fact_assert(c, k, v)` | `fact.assert(c, k, v)` |
| `fact_query(c, k)` | `fact.query(c, k)` |
| `db_connect(dsn)` | `database.connect(dsn)` |
| `db_query(c, sql)` | `database.query(c, sql)` |
| `waf_header(n)` | `waf.header(n)` |
| `waf_cidr_match(ip, cidr)` | `waf.cidr_match(ip, cidr)` |
| `waf_return(action)` | `waf.return(action)` |
