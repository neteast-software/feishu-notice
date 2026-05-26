package feishunotice

import "errors"

// Card is a Feishu interactive card payload.
type Card map[string]any

// Validate checks whether the card can be sent.
func (c Card) Validate() error {
	if len(c) == 0 {
		return errors.New("卡片内容不能为空")
	}
	return nil
}
