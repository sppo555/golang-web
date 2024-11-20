// deck_test.go
package deck

import (
	"testing"
)

func TestNewDeck(t *testing.T) {
	d := newDeck()

	// 檢驗 d 是否為 16
	if len(d) != 16 {
		// 錯誤訊息
		t.Errorf("Expected deck length of 20, but got %v", len(d))
	}
}
