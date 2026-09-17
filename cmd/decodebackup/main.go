package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile("提示词整体备份_2026-7-2.json")
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	text := string(data)
	magic := "XNP1"

	if !strings.HasPrefix(text, magic) {
		fmt.Println("No magic prefix, trying raw JSON")
		var v interface{}
		if err := json.Unmarshal(data, &v); err != nil {
			fmt.Println("JSON parse error:", err)
			return
		}
		fmt.Println("Valid JSON, type:", fmt.Sprintf("%T", v))
		return
	}

	// Decode: remove prefix, shift charCode -3, base64 decode
	shifted := text[len(magic):]
	shiftedBack := make([]byte, len(shifted))
	for i, c := range shifted {
		shiftedBack[i] = byte(c - 3)
	}

	decoded, err := base64.StdEncoding.DecodeString(string(shiftedBack))
	if err != nil {
		fmt.Println("Base64 decode error:", err)
		// Try URL-safe base64
		decoded, err = base64.URLEncoding.DecodeString(string(shiftedBack))
		if err != nil {
			fmt.Println("URL-safe base64 also failed:", err)
			return
		}
	}

	// URI decode
	decodedStr, err := url.PathUnescape(string(decoded))
	if err != nil {
		fmt.Println("URI decode error:", err)
		decodedStr = string(decoded)
	}

	// Parse as JSON
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(decodedStr), &v); err != nil {
		fmt.Println("JSON parse error:", err)
		fmt.Println("First 500 chars:", decodedStr[:min(500, len(decodedStr))])
		return
	}

	fmt.Printf("Type: %v\n", v["type"])
	if dataObj, ok := v["data"].(map[string]interface{}); ok {
		if standalone, ok := dataObj["standalone"].([]interface{}); ok {
			fmt.Printf("Standalone prompts: %d\n", len(standalone))
			if len(standalone) > 0 {
				first := standalone[0].(map[string]interface{})
				fmt.Printf("First prompt keys: ")
				for k := range first {
					fmt.Printf("%s ", k)
				}
				fmt.Println()
				fmt.Printf("First prompt sample: name=%v, category=%v, fields_type=%T, subcategories_type=%T\n",
					first["name"], first["category"], first["fields"], first["subcategories"])
			}
		}
		if packs, ok := dataObj["packs"].(map[string]interface{}); ok {
			fmt.Printf("Packs: %d\n", len(packs))
			// Show first pack's first prompt
			for packName, packData := range packs {
				if prompts, ok := packData.([]interface{}); ok && len(prompts) > 0 {
					fmt.Printf("\nPack: %s (%d prompts)\n", packName, len(prompts))
					first := prompts[0].(map[string]interface{})
					for k, v := range first {
						vtype := fmt.Sprintf("%T", v)
						vstr := fmt.Sprintf("%v", v)
						if len(vstr) > 80 {
							vstr = vstr[:80] + "..."
						}
						fmt.Printf("  %s (%s): %s\n", k, vtype, vstr)
					}
					break
				}
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
