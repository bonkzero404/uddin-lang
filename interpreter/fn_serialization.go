package interpreter

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/goccy/go-json"
)

// valueToJSON converts a Value to JSON bytes
func valueToJSON(v Value) ([]byte, error) {
	switch val := v.(type) {
	case map[string]Value:
		// Convert map to Go map for JSON marshaling
		goMap := make(map[string]any)
		for k, v := range val {
			goMap[k] = valueToInterface(v)
		}
		return json.Marshal(goMap)
	case []Value:
		// Convert array to Go slice for JSON marshaling
		goSlice := make([]any, len(val))
		for i, v := range val {
			goSlice[i] = valueToInterface(v)
		}
		return json.Marshal(goSlice)
	default:
		return json.Marshal(valueToInterface(v))
	}
}

// valueToInterface converts a Value to any for JSON marshaling
func valueToInterface(v Value) any {
	switch val := v.(type) {
	case map[string]Value:
		goMap := make(map[string]any)
		for k, v := range val {
			goMap[k] = valueToInterface(v)
		}
		return goMap
	case []Value:
		goSlice := make([]any, len(val))
		for i, v := range val {
			goSlice[i] = valueToInterface(v)
		}
		return goSlice
	case nil:
		return nil
	default:
		return val
	}
}

// jsonToValue converts JSON any to Value
func jsonToValue(data any) Value {
	switch val := data.(type) {
	case map[string]any:
		result := make(map[string]Value)
		for k, v := range val {
			result[k] = jsonToValue(v)
		}
		return Value(result)
	case []any:
		result := make([]Value, len(val))
		for i, v := range val {
			result[i] = jsonToValue(v)
		}
		return Value(&result)
	case float64:
		// Check if it's actually an integer
		if val == float64(int(val)) {
			return Value(int(val))
		}
		return Value(val)
	case bool:
		return Value(val)
	case string:
		return Value(val)
	case nil:
		return Value(val)
	default:
		return Value(val)
	}
}

// jsonParseFunc implements the json_parse() built-in function
// Parses a JSON string and returns the corresponding Uddin-Lang value
// Parameters:
//   - jsonStr: JSON string to parse
//
// Returns the parsed value (object, array, string, int, float, bool, or null)
// Example: json_parse('{"name": "John", "age": 30}') -> {name: "John", age: 30}
func jsonParseFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "json_parse", args, 1); err != nil {
		return Value(err)
	}

	jsonStr, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "json_parse() requires a string argument, not %s", typeName(args[0])))
	}

	var jsonData any
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		panic(valueError(pos, "Invalid JSON: %v", err))
	}

	return jsonToValue(jsonData)
}

// jsonStringifyFunc implements the json_stringify() built-in function
// Converts a Uddin-Lang value to a JSON string
// Parameters:
//   - value: Value to convert to JSON
//
// Returns the JSON string representation
// Example: json_stringify({name: "John", age: 30}) -> '{"age":30,"name":"John"}'
func jsonStringifyFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "json_stringify", args, 1); err != nil {
		return Value(err)
	}

	jsonBytes, err := valueToJSON(args[0])
	if err != nil {
		panic(valueError(pos, "Failed to convert to JSON: %v", err))
	}

	return Value(string(jsonBytes))
}

// xmlParseFunc implements the xml_parse() built-in function
// Parses an XML string and returns the corresponding Uddin-Lang value
// Parameters:
//   - xmlStr: XML string to parse
//
// Returns the parsed value as a map with XML structure
// Example: xml_parse('<root><name>John</name><age>30</age></root>') -> {root: {name: "John", age: "30"}}
func xmlParseFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "xml_parse", args, 1); err != nil {
		return Value(err)
	}

	xmlStr, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "xml_parse() requires a string argument, not %s", typeName(args[0])))
	}

	// Parse XML into a generic interface
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil // Simple charset handling
	}

	// Parse the XML into a map structure
	result, err := parseXMLElement(decoder)
	if err != nil {
		panic(valueError(pos, "Invalid XML: %v", err))
	}

	return xmlToValue(result)
}

// xmlStringifyFunc implements the xml_stringify() built-in function
// Converts a Uddin-Lang value to an XML string
// Parameters:
//   - value: Value to convert to XML
//
// Returns the XML string representation
// Example: xml_stringify({root: {name: "John", age: "30"}}) -> '<root><name>John</name><age>30</age></root>'
func xmlStringifyFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "xml_stringify", args, 1); err != nil {
		return Value(err)
	}

	xmlBytes, err := valueToXML(args[0])
	if err != nil {
		panic(valueError(pos, "Failed to convert to XML: %v", err))
	}

	return Value(string(xmlBytes))
}

// parseXMLElement parses a single XML element and its children
func parseXMLElement(decoder *xml.Decoder) (map[string]any, error) {
	result := make(map[string]any)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			// Handle start element
			childResult, err := parseXMLElementContent(decoder, t)
			if err != nil {
				return nil, err
			}

			// Handle multiple elements with the same name
			if existing, exists := result[t.Name.Local]; exists {
				// Convert to array if not already
				if arr, isArray := existing.([]any); isArray {
					result[t.Name.Local] = append(arr, childResult)
				} else {
					result[t.Name.Local] = []any{existing, childResult}
				}
			} else {
				result[t.Name.Local] = childResult
			}
		}
	}

	return result, nil
}

// parseXMLElementContent parses the content of an XML element
func parseXMLElementContent(decoder *xml.Decoder, startElement xml.StartElement) (any, error) {
	content := make(map[string]any)
	textContent := ""
	hasChildren := false

	// Add attributes if any
	if len(startElement.Attr) > 0 {
		attrs := make(map[string]any)
		for _, attr := range startElement.Attr {
			attrs[attr.Name.Local] = attr.Value
		}
		content["@attributes"] = attrs
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			hasChildren = true
			childResult, err := parseXMLElementContent(decoder, t)
			if err != nil {
				return nil, err
			}

			// Handle multiple elements with the same name
			if existing, exists := content[t.Name.Local]; exists {
				if arr, isArray := existing.([]any); isArray {
					content[t.Name.Local] = append(arr, childResult)
				} else {
					content[t.Name.Local] = []any{existing, childResult}
				}
			} else {
				content[t.Name.Local] = childResult
			}

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" {
				textContent += text
			}

		case xml.EndElement:
			if t.Name.Local == startElement.Name.Local {
				// End of this element
				if hasChildren {
					if textContent != "" {
						content["#text"] = textContent
					}
					return content, nil
				} else {
					// Simple text element
					if textContent != "" {
						return textContent, nil
					}
					return content, nil
				}
			}
		}
	}
}

// valueToXML converts a Value to XML bytes
func valueToXML(v Value) ([]byte, error) {
	var buf bytes.Buffer
	err := valueToXMLRecursive(v, "", &buf, 0)
	return buf.Bytes(), err
}

// valueToXMLRecursive recursively converts a Value to XML
func valueToXMLRecursive(v Value, tagName string, buf *bytes.Buffer, depth int) error {
	indent := strings.Repeat("  ", depth)

	switch val := v.(type) {
	case map[string]Value:
		if tagName == "" {
			// Root level - find the first key as root element
			for k, v := range val {
				return valueToXMLRecursive(v, k, buf, depth)
			}
		} else {
			// Build opening tag with attributes if present
			openTag := tagName
			var attributes []string

			// Check for attributes
			if attrs, hasAttrs := val["@attributes"]; hasAttrs {
				if attrsMap, ok := attrs.(map[string]Value); ok {
					for attrName, attrValue := range attrsMap {
						attributes = append(attributes, fmt.Sprintf(`%s="%v"`, attrName, attrValue))
					}
				}
			}

			// Write opening tag
			if len(attributes) > 0 {
				fmt.Fprintf(buf, "%s<%s %s>\n", indent, openTag, strings.Join(attributes, " "))
			} else {
				fmt.Fprintf(buf, "%s<%s>\n", indent, openTag)
			}

			// Process child elements
			for k, v := range val {
				// if k == "@attributes" {
				// 	// Already handled above
				// 	continue
				// }
				if k == "#text" {
					// Handle text content
					if str, ok := v.(string); ok {
						buf.WriteString(str)
					}
				} else {
					err := valueToXMLRecursive(v, k, buf, depth+1)
					if err != nil {
						return err
					}
				}
			}
			fmt.Fprintf(buf, "%s</%s>\n", indent, tagName)
		}

	case []Value:
		// Handle arrays - each element gets the same tag name
		for _, item := range val {
			err := valueToXMLRecursive(item, tagName, buf, depth)
			if err != nil {
				return err
			}
		}

	case *[]Value:
		// Handle pointer to arrays
		for _, item := range *val {
			err := valueToXMLRecursive(item, tagName, buf, depth)
			if err != nil {
				return err
			}
		}

	default:
		// Simple value
		content := fmt.Sprintf("%v", val)
		if content == "<nil>" {
			content = ""
		}
		fmt.Fprintf(buf, "%s<%s>%s</%s>\n", indent, tagName, content, tagName)
	}

	return nil
}

// xmlToValue converts XML any to Value
func xmlToValue(data any) Value {
	switch val := data.(type) {
	case map[string]any:
		result := make(map[string]Value)
		for k, v := range val {
			result[k] = xmlToValue(v)
		}
		return Value(result)
	case []any:
		result := make([]Value, len(val))
		for i, v := range val {
			result[i] = xmlToValue(v)
		}
		return Value(&result)
	case string:
		return Value(val)
	case nil:
		return Value(nil)
	default:
		return Value(fmt.Sprintf("%v", val))
	}
}
