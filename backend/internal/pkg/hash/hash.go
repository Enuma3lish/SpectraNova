package hash
package hash














}	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))func ComparePassword(hash, password string) error {}	return string(bytes), nil	}		return "", err	if err != nil {	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)func HashPassword(password string) (string, error) {import "golang.org/x/crypto/bcrypt"