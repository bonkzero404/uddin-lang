package interpreter

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// xmlParseFunc implements the xml_parse() built-in function
func xmlParseFunc(interp *interpreter, pos Position, args []Value) Value {
	if err := ensureNumArgs(pos, "xml_parse", args, 1); err != nil {
		return Value(err)
	}

	xmlStr, ok := args[0].(string)
	if !ok {
		panic(typeError(pos, "xml_parse() requires a string argument, not %s", typeName(args[0])))
	}

	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	result, err := parseXMLElement(decoder)
	if err != nil {
		panic(valueError(pos, "Invalid XML: %v", err))
	}

	return xmlToValue(result)
}

// xmlStringifyFunc implements the xml_stringify() built-in function
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
			childResult, err := parseXMLElementContent(decoder, t)
			if err != nil {
				return nil, err
			}

			if existing, exists := result[t.Name.Local]; exists {
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

func parseXMLElementContent(decoder *xml.Decoder, startElement xml.StartElement) (any, error) {
	content := make(map[string]any)
	textContent := ""
	hasChildren := false

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
				if hasChildren {
					if textContent != "" {
						content["#text"] = textContent
					}
					return content, nil
				}
				if textContent != "" {
					return textContent, nil
				}
				return content, nil
			}
		}
	}
}

func valueToXML(v Value) ([]byte, error) {
	var buf bytes.Buffer
	err := valueToXMLRecursive(v, "", &buf, 0)
	return buf.Bytes(), err
}

func valueToXMLRecursive(v Value, tagName string, buf *bytes.Buffer, depth int) error {
	indent := strings.Repeat("  ", depth)

	switch val := v.(type) {
	case map[string]Value:
		if tagName == "" {
			for k, v := range val {
				return valueToXMLRecursive(v, k, buf, depth)
			}
		} else {
			openTag := tagName
			var attributes []string

			if attrs, hasAttrs := val["@attributes"]; hasAttrs {
				if attrsMap, ok := attrs.(map[string]Value); ok {
					for attrName, attrValue := range attrsMap {
						attributes = append(attributes, fmt.Sprintf(`%s="%v"`, attrName, attrValue))
					}
				}
			}

			if len(attributes) > 0 {
				fmt.Fprintf(buf, "%s<%s %s>\n", indent, openTag, strings.Join(attributes, " "))
			} else {
				fmt.Fprintf(buf, "%s<%s>\n", indent, openTag)
			}

			for k, v := range val {
				if k == "#text" {
					if str, ok := v.(string); ok {
						buf.WriteString(str)
					}
				} else {
					if err := valueToXMLRecursive(v, k, buf, depth+1); err != nil {
						return err
					}
				}
			}
			fmt.Fprintf(buf, "%s</%s>\n", indent, tagName)
		}

	case []Value:
		for _, item := range val {
			if err := valueToXMLRecursive(item, tagName, buf, depth); err != nil {
				return err
			}
		}

	case *[]Value:
		for _, item := range *val {
			if err := valueToXMLRecursive(item, tagName, buf, depth); err != nil {
				return err
			}
		}

	default:
		content := fmt.Sprintf("%v", val)
		if content == "<nil>" {
			content = ""
		}
		fmt.Fprintf(buf, "%s<%s>%s</%s>\n", indent, tagName, content, tagName)
	}

	return nil
}

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
