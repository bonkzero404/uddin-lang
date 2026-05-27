# Syntax Highlighting untuk Uddin-Lang

Dokumentasi ini telah dikonfigurasi untuk mendukung syntax highlighting untuk kode Uddin-Lang dalam blok kode markdown.

## Penggunaan

Untuk menggunakan syntax highlighting Uddin-Lang, gunakan salah satu dari identifier berikut dalam blok kode markdown:

### Menggunakan `uddin`

````markdown
```uddin
func fibonacci(n)
    if n <= 1 then
        return n
    else
        return fibonacci(n-1) + fibonacci(n-2)
    end
end

print(fibonacci(10))
```
````

### Menggunakan `din` (ekstensi file)

````markdown
```din
func greet(name)
    message = "Hello, " + name + "!"
    print(message)
end

greet("World")
```
````

## Fitur Syntax Highlighting

Syntax highlighting Uddin-Lang mendukung:

### Keywords
- `func`, `end`, `if`, `then`, `else`, `elif`
- `while`, `for`, `do`, `break`, `continue`
- `return`, `try`, `catch`, `import`, `as`
- `true`, `false`, `null`

### Built-in Functions
- String functions: `print`, `println`, `len`, `str`, `split`, `join`, `trim`, dll.
- Math functions: `abs`, `ceil`, `floor`, `round`, `sqrt`, `pow`, dll.
- Array functions: `append`, `filter`, `map`, `reduce`, `sort`, dll.
- Date/Time functions: `datetime.now`, `datetime.format`, `datetime.parse`, dll. (requires `import "datetime"`)
- Regex functions: `regex.match`, `regex.find`, `regex.is_match`, dll. (requires `import "regex"`)
- Network functions: `http.get`, `http.post`, `http.tcp_connect`, dll. (requires `import "http"`)
- File functions: `fs.read`, `fs.write`, `fs.exists`, dll. (requires `import "fs"`)

### Lainnya
- **Strings**: Teks dalam tanda kutip ganda
- **Numbers**: Angka integer dan float
- **Comments**: Komentar dengan `//`
- **Operators**: Operator matematika dan logika
- **Variables**: Identifier variabel dan fungsi

## Contoh Lengkap

```uddin
// Contoh program Uddin-Lang dengan syntax highlighting
func calculateFactorial(n)
    if n <= 1 then
        return 1
    else
        return n * calculateFactorial(n - 1)
    end
end

// Main program
numbers = [1, 2, 3, 4, 5]
results = []

for num in numbers do
    factorial = calculateFactorial(num)
    results = append(results, factorial)
    print("Factorial of " + str(num) + " is " + str(factorial))
end

print("All results: " + str(results))
```

## Kustomisasi

Jika Anda ingin memodifikasi warna atau styling syntax highlighting, Anda dapat mengedit file:
- `/src/theme/prism-uddin.js` - Definisi bahasa Prism
- `/src/css/prism-uddin.css` - Styling CSS untuk syntax highlighting

## Troubleshooting

Jika syntax highlighting tidak berfungsi:

1. Pastikan Anda menggunakan identifier yang benar (`uddin` atau `din`)
2. Restart development server dengan `npm start`
3. Clear cache browser Anda
4. Periksa console browser untuk error JavaScript

Jika masih ada masalah, silakan buat issue di repository GitHub.