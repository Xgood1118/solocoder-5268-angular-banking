package masking

import "strings"

func MaskCardNumber(cardNumber string) string {
	if len(cardNumber) <= 4 {
		return cardNumber
	}
	prefix := strings.Repeat("*", len(cardNumber)-4)
	suffix := cardNumber[len(cardNumber)-4:]
	return prefix + suffix
}

func MaskPhone(phone string) string {
	if len(phone) <= 4 {
		return phone
	}
	prefix := phone[:3]
	suffix := phone[len(phone)-4:]
	middle := strings.Repeat("*", len(phone)-7)
	return prefix + middle + suffix
}

func MaskIDCard(idCard string) string {
	if len(idCard) <= 4 {
		return idCard
	}
	prefix := idCard[:4]
	suffix := idCard[len(idCard)-4:]
	middle := strings.Repeat("*", len(idCard)-8)
	return prefix + middle + suffix
}

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}
	username := parts[0]
	domain := parts[1]
	if len(username) <= 2 {
		return username + "@" + domain
	}
	masked := username[:2] + strings.Repeat("*", len(username)-2)
	return masked + "@" + domain
}
