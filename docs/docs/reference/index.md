---
id: index
title: Language Reference
sidebar_label: Reference
---

# Uddin Language Reference

Welcome to the Uddin Programming Language reference documentation. This section provides comprehensive information about the language syntax, built-in functions, and core features.

## Quick Navigation

### [Syntax Reference](./syntax)
Complete guide to Uddin language syntax including:
- Variables and data types
- Control flow statements
- Function definitions
- Operators and expressions
- Module system

### [Module System](./modules)
Reference for the stdlib module system including:
- Import syntax (`import "module"` / `import "module" as alias`)
- All 9 built-in modules: regex, datetime, json, fs, http, waf, cdc, fact, database
- Function reference for each module
- Migration guide from flat global names (pre-v1.2.0)

### [Built-in Functions](./builtin-functions)
Documentation for all built-in functions including:
- String manipulation functions
- Array operations
- Mathematical functions
- I/O operations
- Utility functions

### [Memory Optimization](./memory-optimization)
Production-ready memory layout optimization features including:
- Tagged value types for reduced memory footprint
- Compact environment for thread-safe variable storage
- Cache-friendly structures for large arrays and maps
- Memory leak detection for long-running applications
- Performance optimization techniques and best practices

### [Database Streaming](./database-streaming)
Reference guide for real-time database streaming capabilities including:
- Stream functions and parameters
- Connection management
- Event handling patterns
- Multi-table streaming
- Performance considerations



### [Bytecode VM & Performance](./bytecode-vm)
Internals of the bytecode virtual machine and performance APIs including:
- VM architecture and opcode reference
- `CALL_BUILTIN_DIRECT` fast-path builtins
- Engine program cache and VM pool
- Regex pre-compilation
- Disassembling bytecode

### [Go Library](./library)
Comprehensive guide to using UDDIN-LANG as a Go library including:
- Installation and setup
- Basic usage examples
- API reference
- Integration patterns
- Performance optimization tips

## Language Overview

Uddin is a modern, functional programming language designed for simplicity and expressiveness. Key features include:

- **Clean Syntax**: Easy to read and write
- **Functional Programming**: First-class functions and immutable data structures
- **Built-in Libraries**: Rich set of built-in functions for common tasks
- **Type Safety**: Strong type system with type inference
- **Module System**: Organized code structure with import/export capabilities

## Getting Help

If you need help or have questions about the language:

- Check the [Tutorial](../intro) for step-by-step guides
- Browse [Examples](../examples) for practical code samples
- Visit our [GitHub repository](https://github.com/bonkzero404/uddin-lang) for issues and discussions

---

*This reference documentation is continuously updated. If you find any errors or have suggestions for improvement, please contribute to our documentation.*