package initializers

import "encoding/json"

// JSONMarshal は、構造体をJSON形式にマーシャリングするヘルパー関数です。
func JSONMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
