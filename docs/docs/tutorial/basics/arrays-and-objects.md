---
sidebar_position: 7
---

# Arrays and Objects

Arrays and objects are essential data structures in Uddin-Lang that allow you to store and organize multiple values efficiently.

## Arrays

### Creating Arrays

Arrays store multiple values in a single variable:

```uddin
fun main():
    // Empty array
    empty_array = []

    // Array with numbers
    numbers = [1, 2, 3, 4, 5]

    // Array with strings
    names = ["Alice", "Bob", "Charlie"]

    // Mixed data types
    mixed = [1, "hello", true, 3.14]

    print("Numbers: " + str(numbers))
    print("Names: " + str(names))
end
```

### Accessing Array Elements

Use square brackets with an index (starting from 0):

```uddin
fun main():
    fruits = ["apple", "banana", "orange", "grape"]

    // Access elements by index
    first_fruit = fruits[0]    // "apple"
    second_fruit = fruits[1]   // "banana"
    last_fruit = fruits[3]     // "grape"

    print("First fruit: " + first_fruit)
    print("Second fruit: " + second_fruit)
    print("Last fruit: " + last_fruit)
end
```

### Modifying Arrays

```uddin
fun main():
    numbers = [10, 20, 30]
    print("Original: " + str(numbers))

    // Change an element
    numbers[1] = 25
    print("After change: " + str(numbers))

    // Add elements using push
    push(numbers, 40)
    push(numbers, 50)
    print("After adding: " + str(numbers))
end
```

### Array Length

```uddin
fun main():
    colors = ["red", "green", "blue"]

    array_length = len(colors)
    print("Array has " + str(array_length) + " elements")

    // Check if array is empty
    if (len(colors) == 0) then:
        print("Array is empty")
    else:
        print("Array has elements")
    end
end
```

### Iterating Through Arrays

#### Using For Loops

```uddin
fun main():
    scores = [85, 92, 78, 96, 88]

    print("All scores:")
    for (score in scores):
        print("Score: " + str(score))
    end
end
```

#### Using While Loops

```uddin
fun main():
    names = ["Alice", "Bob", "Charlie"]

    i = 0
    while (i < len(names)):
        print("Name " + str(i + 1) + ": " + names[i])
        i = i + 1
    end
end
```

### Array Operations

#### Finding Elements

```uddin
fun find_element(array, target):
    i = 0
    while (i < len(array)):
        if (array[i] == target) then:
            return i
        end
        i = i + 1
    end
    return -1  // Not found
end

fun main():
    numbers = [10, 25, 30, 45, 50]
    target = 30

    index = find_element(numbers, target)
    if (index != -1) then:
        print("Found " + str(target) + " at index " + str(index))
    else:
        print(str(target) + " not found")
    end
end
```

#### Calculating Statistics

```uddin
fun calculate_sum(numbers):
    total = 0
    for (num in numbers):
        total = total + num
    end
    return total
end

fun find_maximum(numbers):
    if (len(numbers) == 0) then:
        return null
    end

    max_val = numbers[0]
    for (num in numbers):
        if (num > max_val) then:
            max_val = num
        end
    end
    return max_val
end

fun main():
    data = [12, 5, 8, 15, 3, 9]

    sum = calculate_sum(data)
    max_val = find_maximum(data)
    average = sum / len(data)

    print("Data: " + str(data))
    print("Sum: " + str(sum))
    print("Maximum: " + str(max_val))
    print("Average: " + str(average))
end
```

## Objects

### Creating Objects

Objects store key-value pairs:

```uddin
fun main():
    // Empty object
    empty_obj = {}

    // Object with properties
    person = {
        name: "Alice",
        age: 25,
        city: "New York"
    }

    // Object with mixed data types
    product = {
        id: 1001,
        name: "Laptop",
        price: 999.99,
        in_stock: true,
        tags: ["electronics", "computer"]
    }

    print("Person: " + str(person))
    print("Product: " + str(product))
end
```

### Accessing Object Properties

Use dot notation or bracket notation:

```uddin
fun main():
    student = {
        name: "Bob",
        grade: 85,
        subject: "Math"
    }

    // Dot notation
    student_name = student.name
    student_grade = student.grade

    // Bracket notation
    student_subject = student["subject"]

    print("Student: " + student_name)
    print("Grade: " + str(student_grade))
    print("Subject: " + student_subject)
end
```

### Modifying Objects

```uddin
fun main():
    car = {
        brand: "Toyota",
        model: "Camry",
        year: 2020
    }

    print("Original: " + str(car))

    // Modify existing property
    car.year = 2021

    // Add new property
    car.color = "Blue"
    car.mileage = 15000

    print("Modified: " + str(car))
end
```

### Object Methods

```uddin
fun create_person(name, age, city):
    return {
        name: name,
        age: age,
        city: city
    }
end

fun get_person_info(person):
    return person.name + " is " + str(person.age) + " years old and lives in " + person.city
end

fun main():
    person1 = create_person("Alice", 25, "New York")
    person2 = create_person("Bob", 30, "Los Angeles")

    print(get_person_info(person1))
    print(get_person_info(person2))
end
```

## Nested Data Structures

### Arrays of Objects

```uddin
fun main():
    students = [
        {name: "Alice", grade: 85, subject: "Math"},
        {name: "Bob", grade: 92, subject: "Science"},
        {name: "Charlie", grade: 78, subject: "History"}
    ]

    print("Student grades:")
    for (student in students):
        print(student.name + ": " + str(student.grade) + " in " + student.subject)
    end

    // Calculate average grade
    total = 0
    for (student in students):
        total = total + student.grade
    end
    average = total / len(students)
    print("Average grade: " + str(average))
end
```

### Objects with Arrays

```uddin
fun main():
    classroom = {
        teacher: "Ms. Johnson",
        subject: "Mathematics",
        students: ["Alice", "Bob", "Charlie", "Diana"],
        grades: [85, 92, 78, 88]
    }

    print("Class: " + classroom.subject)
    print("Teacher: " + classroom.teacher)
    print("Number of students: " + str(len(classroom.students)))

    // Print student grades
    i = 0
    while (i < len(classroom.students)):
        student = classroom.students[i]
        grade = classroom.grades[i]
        print(student + ": " + str(grade))
        i = i + 1
    end
end
```

### Nested Objects

```uddin
fun main():
    company = {
        name: "Tech Corp",
        address: {
            street: "123 Main St",
            city: "San Francisco",
            state: "CA",
            zip: "94105"
        },
        employees: [
            {
                name: "Alice",
                position: "Developer",
                contact: {
                    email: "alice@techcorp.com",
                    phone: "555-1234"
                }
            },
            {
                name: "Bob",
                position: "Designer",
                contact: {
                    email: "bob@techcorp.com",
                    phone: "555-5678"
                }
            }
        ]
    }

    print("Company: " + company.name)
    print("Location: " + company.address.city + ", " + company.address.state)

    print("\nEmployees:")
    for (employee in company.employees):
        print(employee.name + " - " + employee.position)
        print("Email: " + employee.contact.email)
    end
end
```

## Practical Examples

### Example 1: Student Grade Manager

```uddin
fun add_student(students, name, grades):
    student = {
        name: name,
        grades: grades,
        average: calculate_average(grades)
    }
    push(students, student)
end

fun calculate_average(grades):
    if (len(grades) == 0) then:
        return 0
    end

    total = 0
    for (grade in grades):
        total = total + grade
    end
    return total / len(grades)
end

fun find_top_student(students):
    if (len(students) == 0) then:
        return null
    end

    top_student = students[0]
    for (student in students):
        if (student.average > top_student.average) then:
            top_student = student
        end
    end
    return top_student
end

fun main():
    students = []

    // Add students
    add_student(students, "Alice", [85, 92, 88, 90])
    add_student(students, "Bob", [78, 85, 82, 79])
    add_student(students, "Charlie", [95, 98, 92, 96])

    // Display all students
    print("All students:")
    for (student in students):
        print(student.name + " - Average: " + str(student.average))
    end

    // Find top student
    top = find_top_student(students)
    print("\nTop student: " + top.name + " with average " + str(top.average))
end
```

### Example 2: Inventory Management

```uddin
fun create_product(id, name, price, quantity):
    return {
        id: id,
        name: name,
        price: price,
        quantity: quantity
    }
end

fun add_product(inventory, product):
    push(inventory, product)
end

fun find_product(inventory, id):
    for (product in inventory):
        if (product.id == id) then:
            return product
        end
    end
    return null
end

fun calculate_total_value(inventory):
    total = 0
    for (product in inventory):
        total = total + (product.price * product.quantity)
    end
    return total
end

fun main():
    inventory = []

    // Add products
    add_product(inventory, create_product(1, "Laptop", 999.99, 10))
    add_product(inventory, create_product(2, "Mouse", 25.50, 50))
    add_product(inventory, create_product(3, "Keyboard", 75.00, 25))

    // Display inventory
    print("Inventory:")
    for (product in inventory):
        print("ID: " + str(product.id) + ", Name: " + product.name + ", Price: $" + str(product.price) + ", Qty: " + str(product.quantity))
    end

    // Calculate total value
    total_value = calculate_total_value(inventory)
    print("\nTotal inventory value: $" + str(total_value))

    // Find specific product
    product = find_product(inventory, 2)
    if (product != null) then:
        print("\nFound product: " + product.name)
    end
end
```

## Best Practices

### 1. Use Meaningful Names

```uddin
// Good: Descriptive property names
user = {
    first_name: "John",
    last_name: "Doe",
    email_address: "john@example.com",
    birth_date: "1990-01-01"
}

// Avoid: Unclear names
user = {
    fn: "John",
    ln: "Doe",
    em: "john@example.com",
    bd: "1990-01-01"
}
```

### 2. Keep Data Structures Simple

```uddin
// Good: Simple, flat structure when possible
product = {
    name: "Laptop",
    price: 999.99,
    category: "Electronics"
}

// Avoid: Unnecessary nesting
product = {
    details: {
        info: {
            name: "Laptop"
        },
        pricing: {
            amount: 999.99
        }
    }
}
```

### 3. Validate Data

```uddin
fun create_user(name, age, email):
    if (len(name) == 0) then:
        print("Error: Name cannot be empty")
        return null
    end

    if (age < 0) then:
        print("Error: Age cannot be negative")
        return null
    end

    return {
        name: name,
        age: age,
        email: email
    }
end
```

### 4. Use Helper Functions

```uddin
fun is_valid_email(email):
    return contains(email, "@") && contains(email, ".")
end

fun format_name(first, last):
    return first + " " + last
end

fun create_contact(first_name, last_name, email):
    if (!is_valid_email(email)) then:
        print("Error: Invalid email")
        return null
    end

    return {
        name: format_name(first_name, last_name),
        email: email
    }
end
```

## Practice Exercises

### Exercise 1: Shopping Cart

Create a shopping cart system:

```uddin
fun add_item(cart, name, price, quantity):
    // Add item to cart
end

fun remove_item(cart, name):
    // Remove item from cart
end

fun calculate_total(cart):
    // Calculate total price
end

fun main():
    cart = []

    add_item(cart, "Apple", 1.50, 5)
    add_item(cart, "Bread", 2.00, 1)
    add_item(cart, "Milk", 3.50, 2)

    print("Total: $" + str(calculate_total(cart)))
end
```

### Exercise 2: Library System

Create a simple library management system:

```uddin
fun add_book(library, title, author, year):
    // Add book to library
end

fun find_books_by_author(library, author):
    // Return array of books by specific author
end

fun count_books(library):
    // Return total number of books
end

fun main():
    library = []

    add_book(library, "1984", "George Orwell", 1949)
    add_book(library, "Animal Farm", "George Orwell", 1945)
    add_book(library, "To Kill a Mockingbird", "Harper Lee", 1960)

    orwell_books = find_books_by_author(library, "George Orwell")
    print("Books by George Orwell: " + str(len(orwell_books)))
end
```

### Exercise 3: Address Book

Create an address book application:

```uddin
fun add_contact(contacts, name, phone, email, address):
    // Add contact to address book
end

fun find_contact(contacts, name):
    // Find contact by name
end

fun list_contacts(contacts):
    // Print all contacts
end

fun main():
    contacts = []

    add_contact(contacts, "Alice", "555-1234", "alice@email.com", "123 Main St")
    add_contact(contacts, "Bob", "555-5678", "bob@email.com", "456 Oak Ave")

    list_contacts(contacts)
end
```

## Summary

Arrays and objects are powerful data structures that enable you to:

-   **Store multiple values** efficiently in arrays
-   **Organize related data** using objects
-   **Create complex data models** with nested structures
-   **Build real-world applications** like inventory systems, user management, and more

Mastering these data structures is essential for building sophisticated Uddin-Lang applications.

---

**Next**: Learn about [Error Handling](error-handling) to make your programs more robust and reliable.
