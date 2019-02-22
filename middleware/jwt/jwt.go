package jwt

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	gojwt "github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
)

// errorHandler handle error and return a bool value
// return false will break conext, return true will handle next context
type errorHandler func(resp *restful.Response, err error) bool

// TokenExtractor extract jwt token
type tokenExtractor func(name string, req *restful.Request) (string, error)

// customValidator jwt standard validator
type customValidator func(config *Config, req *restful.Request) error

// Provider JWT validator handler
type Provider struct {
	Config *Config
}

//Config JWT Middleware config
type Config struct {
	// Name token name, default: Authorization
	Name string
	// Signing key to validate token
	SigningKey string
	// ErrorHandler validate error handler, default: defaultOnError
	ErrorHandler errorHandler
	// Extractor extract jwt token, default extract from header: defaultExtractorFromHeader
	Extractor tokenExtractor
	// EnableAuthOnOptions http option method validate switch
	EnableAuthOnOptions bool
	// SigningMethod sign method, default: HS256
	SigningMethod gojwt.SigningMethod
	// ExcludeURL exclude url will skip jwt validator
	ExcludeURL []string
	// ExcludePrefix exclude url prefix will skip jwt validator
	ExcludePrefix []string
	// ContextKey Context key to store user information from the token into context.
	// Optional. Default value "user".
	ContextKey string

	// CustomValidator custom validator suggestion flow：
	// 1. check exlude url, and exclude url prefix
	// 2. extract token string
	// 3. check token sign
	// 4. check token ttl
	// 5. save custom value to conext after check passed
	// 6. handle addon validator
	CustomValidator customValidator
	// validationKeyGetter
	validationKeyGetter gojwt.Keyfunc
}

// New create a JWT provider
func New(config *Config) *Provider {
	if config.Name == "" {
		config.Name = "Authorization"
	}
	if config.ContextKey == "" {
		config.ContextKey = config.Name
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultOnError
	}
	if config.Extractor == nil {
		config.Extractor = defaultExtractorFromHeader
	}
	if config.SigningMethod == nil {
		config.SigningMethod = gojwt.SigningMethodHS256
	}

	if config.SigningKey == "" {
		panic("jwt middleware requires signing key")
	} else {
		config.validationKeyGetter = func(token *gojwt.Token) (interface{}, error) {
			return []byte(config.SigningKey), nil
		}
	}
	if config.CustomValidator == nil {
		config.CustomValidator = defaultCheckJWT
	}

	return &Provider{config}
}

// GeneratorToken generate token by custom value and token ttl
func (t *Provider) GeneratorToken(customValue string, ttl time.Duration) (string, error) {
	claims := make(gojwt.MapClaims)
	claims[t.Config.ContextKey] = customValue
	claims["exp"] = time.Now().Add(ttl).Unix()
	token := gojwt.NewWithClaims(t.Config.SigningMethod, claims)
	// sign token and get the complete encoded token as a string
	return token.SignedString([]byte(t.Config.SigningKey))
}

// GetCustomValue return custom value in token, or returns error
func (t *Provider) GetCustomValue(req *restful.Request) (string, error) {
	val := req.Attribute(t.Config.ContextKey)
	if val == nil {
		return "", fmt.Errorf("token value not exist")
	}
	return val.(string), nil
}

// defaultOnError default error handler
// return false will break conext, return true will handle next context
func defaultOnError(resp *restful.Response, err error) bool {
	resp.WriteHeader(http.StatusUnauthorized)
	resp.Write([]byte(err.Error()))
	return false
}

// defaultExtractorFromHeader extract token from header
func defaultExtractorFromHeader(name string, req *restful.Request) (string, error) {
	authHeader := req.HeaderParameter(name)
	if authHeader == "" || len(authHeader) <= 7 {
		return "", nil // No error, just no token
	}
	// TODO: Make this a bit more robust, parsing-wise
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", fmt.Errorf("Authorization header format must be Bearer {token}")
	}

	return authHeaderParts[1], nil
}

func (t *Provider) Auth(req *restful.Request, resp *restful.Response,
	chain *restful.FilterChain) {
	if err := t.Config.CustomValidator(t.Config, req); err != nil {
		if ret := t.Config.ErrorHandler(resp, err); ret == false {
			return
		}
	}
	chain.ProcessFilter(req, resp)
}

// defaultCheckJWT execlude token check flow, or returns error
// 1. check exlude url, and exclude url prefix
// 2. extract token string
// 3. check token sign
// 4. check token ttl
// 5. save custom value to conext after check passed
func defaultCheckJWT(config *Config, req *restful.Request) error {
	r := req.Request

	if !config.EnableAuthOnOptions {
		if r.Method == "OPTIONS" {
			return nil
		}
	}

	// check exclude url
	if len(config.ExcludeURL) > 0 {
		for _, url := range config.ExcludeURL {
			if url == r.URL.Path {
				return nil
			}
		}
	}
	// check exclude url prefix
	if len(config.ExcludePrefix) > 0 {
		for _, prefix := range config.ExcludePrefix {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return nil
			}
		}
	}

	// extract token
	token, err := config.Extractor(config.Name, req)

	if err != nil {
		return fmt.Errorf("Error extracting token: %v", err)
	}
	if token == "" {
		// no token
		return fmt.Errorf("Required authorization token not found")
	}

	// parse token value
	parsedToken, err := gojwt.Parse(token, config.validationKeyGetter)

	if err != nil {
		return fmt.Errorf("Error parsing token: %v", err)
	}

	if config.SigningMethod != nil && config.SigningMethod.Alg() != parsedToken.Header["alg"] {
		message := fmt.Sprintf("Expected %s signing method but token specified %s",
			config.SigningMethod.Alg(),
			parsedToken.Header["alg"])
		return fmt.Errorf("Error validating token algorithm: %s", message)
	}

	if !parsedToken.Valid {
		return fmt.Errorf("Token is invalid")
	}
	// save custom value to context
	claims := parsedToken.Claims.(gojwt.MapClaims)
	req.SetAttribute(config.ContextKey, claims[config.ContextKey])

	return nil
}
