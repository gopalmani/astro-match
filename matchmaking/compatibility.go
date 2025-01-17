package matchmaking

import (
	"astromatch/db"
	"math/rand"
)

// CompatibilityMatrix - Zodiac compatibility scores
var CompatibilityMatrix = map[string]map[string]int{
	"Aries": {
		"Aries": 75, "Taurus": 70, "Gemini": 85, "Cancer": 60, "Leo": 90, "Virgo": 65,
		"Libra": 80, "Scorpio": 55, "Sagittarius": 88, "Capricorn": 60, "Aquarius": 77, "Pisces": 50,
	},
	"Taurus": {
		"Aries": 70, "Taurus": 80, "Gemini": 65, "Cancer": 85, "Leo": 55, "Virgo": 88,
		"Libra": 60, "Scorpio": 82, "Sagittarius": 58, "Capricorn": 84, "Aquarius": 55, "Pisces": 75,
	},
	"Gemini": {
		"Aries": 85, "Taurus": 65, "Gemini": 78, "Cancer": 62, "Leo": 80, "Virgo": 67,
		"Libra": 90, "Scorpio": 53, "Sagittarius": 83, "Capricorn": 57, "Aquarius": 88, "Pisces": 60,
	},
	"Cancer": {
		"Aries": 60, "Taurus": 85, "Gemini": 62, "Cancer": 80, "Leo": 70, "Virgo": 75,
		"Libra": 55, "Scorpio": 90, "Sagittarius": 50, "Capricorn": 82, "Aquarius": 52, "Pisces": 88,
	},
	"Leo": {
		"Aries": 90, "Taurus": 55, "Gemini": 80, "Cancer": 70, "Leo": 85, "Virgo": 60,
		"Libra": 87, "Scorpio": 66, "Sagittarius": 92, "Capricorn": 58, "Aquarius": 81, "Pisces": 54,
	},
	"Virgo": {
		"Aries": 65, "Taurus": 88, "Gemini": 67, "Cancer": 75, "Leo": 60, "Virgo": 85,
		"Libra": 70, "Scorpio": 80, "Sagittarius": 55, "Capricorn": 90, "Aquarius": 63, "Pisces": 77,
	},
	"Libra": {
		"Aries": 80, "Taurus": 60, "Gemini": 90, "Cancer": 55, "Leo": 87, "Virgo": 70,
		"Libra": 88, "Scorpio": 59, "Sagittarius": 85, "Capricorn": 65, "Aquarius": 92, "Pisces": 58,
	},
	"Scorpio": {
		"Aries": 55, "Taurus": 82, "Gemini": 53, "Cancer": 90, "Leo": 66, "Virgo": 80,
		"Libra": 59, "Scorpio": 85, "Sagittarius": 57, "Capricorn": 86, "Aquarius": 50, "Pisces": 89,
	},
	"Sagittarius": {
		"Aries": 88, "Taurus": 58, "Gemini": 83, "Cancer": 50, "Leo": 92, "Virgo": 55,
		"Libra": 85, "Scorpio": 57, "Sagittarius": 80, "Capricorn": 60, "Aquarius": 79, "Pisces": 52,
	},
	"Capricorn": {
		"Aries": 60, "Taurus": 84, "Gemini": 57, "Cancer": 82, "Leo": 58, "Virgo": 90,
		"Libra": 65, "Scorpio": 86, "Sagittarius": 60, "Capricorn": 85, "Aquarius": 62, "Pisces": 78,
	},
	"Aquarius": {
		"Aries": 77, "Taurus": 55, "Gemini": 88, "Cancer": 52, "Leo": 81, "Virgo": 63,
		"Libra": 92, "Scorpio": 50, "Sagittarius": 79, "Capricorn": 62, "Aquarius": 87, "Pisces": 55,
	},
	"Pisces": {
		"Aries": 50, "Taurus": 75, "Gemini": 60, "Cancer": 88, "Leo": 54, "Virgo": 77,
		"Libra": 58, "Scorpio": 89, "Sagittarius": 52, "Capricorn": 78, "Aquarius": 55, "Pisces": 85,
	},
}

// GetZodiacCompatibility - Calculates compatibility between two zodiac signs
func GetZodiacCompatibility(sign1, sign2 string) int {
	if score, exists := CompatibilityMatrix[sign1][sign2]; exists {
		return score
	}
	// Default to random compatibility if no match is found
	return rand.Intn(100)
}

func FindCompatibleUsers(user db.User) ([]db.User, error) {
	allUsers, err := db.GetAllUsers()
	if err != nil {
		return nil, err
	}

	var compatibleUsers []db.User
	for _, u := range allUsers {
		if u.ZodiacSign == user.ZodiacSign { // Example compatibility logic
			compatibleUsers = append(compatibleUsers, u)
		}
	}
	return compatibleUsers, nil
}
