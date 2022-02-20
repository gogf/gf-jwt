package jwt

import (
	"context"
	"crypto/rsa"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcache"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// MapClaims type that uses the map[string]interface{} for JSON decoding
// This is the default claims type if you don't supply one
type MapClaims map[string]interface{}

// GfJWTMiddleware provides a Json-Web-Token authentication implementation. On failure, a 401 HTTP response
// is returned. On success, the wrapped middleware is called, and the userID is made available as
// c.Get("userID").(string).
// Users can get a token by posting a json request to LoginHandler. The token then needs to be passed in
// the Authentication header. Example: Authorization:Bearer XXX_TOKEN_XXX
type GfJWTMiddleware struct {
	// Realm name to display to the user. Required.
	Realm string

	// signing algorithm - possible values are HS256, HS384, HS512
	// Optional, default is HS256.
	SigningAlgorithm string

	// Secret key used for signing. Required.
	Key []byte

	// Duration that a jwt token is valid. Optional, defaults to one hour.
	Timeout time.Duration

	// This field allows clients to refresh their token until MaxRefresh has passed.
	// Note that clients can refresh their token in the last moment of MaxRefresh.
	// This means that the maximum validity timespan for a token is TokenTime + MaxRefresh.
	// Optional, defaults to 0 meaning not refreshable.
	MaxRefresh time.Duration

	// Callback function that should perform the authentication of the user based on login info.
	// Must return user data as user identifier, it will be stored in Claim Array. Required.
	// Check error (e) to determine the appropriate error message.
	Authenticator func(r *ghttp.Request) (interface{}, error)

	// Callback function that should perform the authorization of the authenticated user. Called
	// only after an authentication success. Must return true on success, false on failure.
	// Optional, default to success.
	Authorizator func(data interface{}, r *ghttp.Request) bool

	// Callback function that will be called during login.
	// Using this function it is possible to add additional payload data to the webtoken.
	// The data is then made available during requests via c.Get(jwt.PayloadKey).
	// Note that the payload is not encrypted.
	// The attributes mentioned on jwt.io can't be used as keys for the map.
	// Optional, by default no additional data will be set.
	PayloadFunc func(data interface{}) MapClaims

	// User can define own Unauthorized func.
	Unauthorized func(*ghttp.Request, int, string)

	// User can define own LoginResponse func.
	LoginResponse func(*ghttp.Request, int, string, time.Time)

	// User can define own RefreshResponse func.
	RefreshResponse func(*ghttp.Request, int, string, time.Time)

	// User can define own LogoutResponse func.
	LogoutResponse func(*ghttp.Request, int)

	// Set the identity handler function
	IdentityHandler func(*ghttp.Request) interface{}

	// Set the identity key
	IdentityKey string

	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup string

	// TokenHeadName is a string in the header. Default value is "Bearer"
	TokenHeadName string

	// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
	TimeFunc func() time.Time

	// HTTP Status messages for when something in the JWT middleware fails.
	// Check error (e) to determine the appropriate error message.
	HTTPStatusMessageFunc func(e error, r *ghttp.Request) string

	// Private key file for asymmetric algorithms
	PrivKeyFile string

	// Public key file for asymmetric algorithms
	PubKeyFile string

	// Private key
	privKey *rsa.PrivateKey

	// Public key
	pubKey *rsa.PublicKey

	// Optionally return the token as a cookie
	SendCookie bool

	// Allow insecure cookies for development over http
	SecureCookie bool

	// Allow cookies to be accessed client side for development
	CookieHTTPOnly bool

	// Allow cookie domain change for development
	CookieDomain string

	// SendAuthorization allow return authorization header for every request
	SendAuthorization bool

	// Disable abort() of context.
	DisabledAbort bool

	// CookieName allow cookie name change for development
	CookieName string

	// CacheAdapter
	CacheAdapter gcache.Adapter
	// context
	Ctx context.Context
}

var (
	// TokenKey default jwt token key in params
	TokenKey = "JWT_TOKEN"
	// PayloadKey default jwt payload key in params
	PayloadKey = "JWT_PAYLOAD"
	// IdentityKey default identity key
	IdentityKey = "identity"
	// The blacklist stores tokens that have not expired but have been deactivated.
	blacklist = gcache.New()
)

// New for check error with GfJWTMiddleware
func New(m *GfJWTMiddleware) (*GfJWTMiddleware, error) {
	if err := m.MiddlewareInit(); err != nil {
		return nil, err
	}

	return m, nil
}
func (mw *GfJWTMiddleware) SetCtx(ctx context.Context) *GfJWTMiddleware {
	mw.Ctx = ctx
	return mw
}
func (mw *GfJWTMiddleware) readKeys() error {
	err := mw.privateKey()
	if err != nil {
		return err
	}
	err = mw.publicKey()
	if err != nil {
		return err
	}
	return nil
}

func (mw *GfJWTMiddleware) privateKey() error {
	keyData, err := ioutil.ReadFile(mw.PrivKeyFile)
	if err != nil {
		return ErrNoPrivKeyFile
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return ErrInvalidPrivKey
	}
	mw.privKey = key
	return nil
}

func (mw *GfJWTMiddleware) publicKey() error {
	keyData, err := ioutil.ReadFile(mw.PubKeyFile)
	if err != nil {
		return ErrNoPubKeyFile
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return ErrInvalidPubKey
	}
	mw.pubKey = key
	return nil
}

func (mw *GfJWTMiddleware) usingPublicKeyAlgo() bool {
	switch mw.SigningAlgorithm {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

// MiddlewareInit initialize jwt configs.
func (mw *GfJWTMiddleware) MiddlewareInit() error {
	if mw.Ctx == nil {
		return ErrMissingContext
	}
	if mw.TokenLookup == "" {
		mw.TokenLookup = "header:Authorization"
	}

	if mw.SigningAlgorithm == "" {
		mw.SigningAlgorithm = "HS256"
	}

	if mw.Timeout == 0 {
		mw.Timeout = time.Hour
	}

	if mw.TimeFunc == nil {
		mw.TimeFunc = time.Now
	}

	mw.TokenHeadName = strings.TrimSpace(mw.TokenHeadName)
	if len(mw.TokenHeadName) == 0 {
		mw.TokenHeadName = "Bearer"
	}

	if mw.Authorizator == nil {
		mw.Authorizator = func(data interface{}, r *ghttp.Request) bool {
			return true
		}
	}

	if mw.Unauthorized == nil {
		mw.Unauthorized = func(r *ghttp.Request, code int, message string) {
			r.Response.WriteJson(g.Map{
				"code":    code,
				"message": message,
			})
		}
	}

	if mw.LoginResponse == nil {
		mw.LoginResponse = func(r *ghttp.Request, code int, token string, expire time.Time) {
			r.Response.WriteJson(g.Map{
				"code":   http.StatusOK,
				"token":  token,
				"expire": expire.Format(time.RFC3339),
			})
		}
	}

	if mw.RefreshResponse == nil {
		mw.RefreshResponse = func(r *ghttp.Request, code int, token string, expire time.Time) {
			r.Response.WriteJson(g.Map{
				"code":   http.StatusOK,
				"token":  token,
				"expire": expire.Format(time.RFC3339),
			})
		}
	}

	if mw.LogoutResponse == nil {
		mw.LogoutResponse = func(r *ghttp.Request, code int) {
			r.Response.WriteJson(g.Map{
				"code":    http.StatusOK,
				"message": "success",
			})
		}
	}

	if mw.IdentityKey == "" {
		mw.IdentityKey = IdentityKey
	}

	if mw.IdentityHandler == nil {
		mw.IdentityHandler = func(r *ghttp.Request) interface{} {
			claims := ExtractClaims(r)
			return claims[mw.IdentityKey]
		}
	}

	if mw.HTTPStatusMessageFunc == nil {
		mw.HTTPStatusMessageFunc = func(e error, r *ghttp.Request) string {
			return e.Error()
		}
	}

	if mw.Realm == "" {
		mw.Realm = "gf jwt"
	}

	if mw.CookieName == "" {
		mw.CookieName = "jwt"
	}

	if mw.usingPublicKeyAlgo() {
		return mw.readKeys()
	}

	if mw.Key == nil {
		return ErrMissingSecretKey
	}

	if mw.CacheAdapter != nil {
		blacklist.SetAdapter(mw.CacheAdapter)
	}

	return nil
}

// MiddlewareFunc makes GfJWTMiddleware implement the Middleware interface.
func (mw *GfJWTMiddleware) MiddlewareFunc() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		mw.middlewareImpl(r)
	}
}

func (mw *GfJWTMiddleware) middlewareImpl(r *ghttp.Request) {
	claims, token, err := mw.GetClaimsFromJWT(r)
	if err != nil {
		mw.unauthorized(r, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, r))
		return
	}

	if claims["exp"] == nil {
		mw.unauthorized(r, http.StatusBadRequest, mw.HTTPStatusMessageFunc(ErrMissingExpField, r))
		return
	}

	if _, ok := claims["exp"].(float64); !ok {
		mw.unauthorized(r, http.StatusBadRequest, mw.HTTPStatusMessageFunc(ErrWrongFormatOfExp, r))
		return
	}

	if int64(claims["exp"].(float64)) < mw.TimeFunc().Unix() {
		mw.unauthorized(r, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(ErrExpiredToken, r))
		return
	}

	in, err := mw.inBlacklist(token)
	if err != nil {
		mw.unauthorized(r, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, r))
		return
	}

	if in {
		mw.unauthorized(r, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(ErrInvalidToken, r))
		return
	}

	r.SetParam(PayloadKey, claims)
	identity := mw.IdentityHandler(r)

	if identity != nil {
		r.SetParam(mw.IdentityKey, identity)
	}

	if !mw.Authorizator(identity, r) {
		mw.unauthorized(r, http.StatusForbidden, mw.HTTPStatusMessageFunc(ErrForbidden, r))
		return
	}

	//c.Next() todo
}

// GetClaimsFromJWT get claims from JWT token
func (mw *GfJWTMiddleware) GetClaimsFromJWT(r *ghttp.Request) (MapClaims, string, error) {
	token, err := mw.ParseToken(r)

	if err != nil {
		return nil, "", err
	}

	if mw.SendAuthorization {
		token := r.Get(TokenKey).String()
		if len(token) > 0 {
			r.Header.Set("Authorization", mw.TokenHeadName+" "+token)
		}
	}

	claims := MapClaims{}
	for key, value := range token.Claims.(jwt.MapClaims) {
		claims[key] = value
	}

	return claims, token.Raw, nil
}

// LoginHandler can be used by clients to get a jwt token.
// Payload needs to be json in the form of {"username": "USERNAME", "password": "PASSWORD"}.
// Reply will be of the form {"token": "TOKEN"}.
func (mw *GfJWTMiddleware) LoginHandler(r *ghttp.Request) {
	if mw.Authenticator == nil {
		mw.unauthorized(r, http.StatusInternalServerError, mw.HTTPStatusMessageFunc(ErrMissingAuthenticatorFunc, r))
		return
	}

	data, err := mw.Authenticator(r)

	if err != nil {
		mw.unauthorized(r, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, r))
		return
	}

	// Create the token
	token := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))
	claims := token.Claims.(jwt.MapClaims)

	if mw.PayloadFunc != nil {
		for key, value := range mw.PayloadFunc(data) {
			claims[key] = value
		}
	}

	if _, ok := claims[mw.IdentityKey]; !ok {
		mw.unauthorized(r, http.StatusInternalServerError, mw.HTTPStatusMessageFunc(ErrMissingIdentity, r))
		return
	}

	expire := mw.TimeFunc().Add(mw.Timeout)
	claims["exp"] = expire.Unix()
	claims["iat"] = mw.TimeFunc().Unix()
	tokenString, err := mw.signedString(token)

	if err != nil {
		mw.unauthorized(r, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(ErrFailedTokenCreation, r))
		return
	}

	// set cookie
	if mw.SendCookie {
		maxage := int64(expire.Unix() - time.Now().Unix())
		r.Cookie.SetCookie(mw.CookieName, tokenString, mw.CookieDomain, "/", time.Duration(maxage)*time.Second)
	}

	mw.LoginResponse(r, http.StatusOK, tokenString, expire)
}

func (mw *GfJWTMiddleware) signedString(token *jwt.Token) (string, error) {
	var tokenString string
	var err error
	if mw.usingPublicKeyAlgo() {
		tokenString, err = token.SignedString(mw.privKey)
	} else {
		tokenString, err = token.SignedString(mw.Key)
	}
	return tokenString, err
}

// LogoutHandler can be used to logout a token. The token still needs to be valid on logout.
// Logout the token puts the unexpired token on a blacklist.
func (mw *GfJWTMiddleware) LogoutHandler(r *ghttp.Request) {
	claims, token, err := mw.CheckIfTokenExpire(r)
	if err != nil {
		mw.unauthorized(r, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, r))
		return
	}

	err = mw.setBlacklist(token, claims)

	if err != nil {
		mw.unauthorized(r, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, r))
		return
	}

	mw.LogoutResponse(r, http.StatusOK)
}

// RefreshHandler can be used to refresh a token. The token still needs to be valid on refresh.
// Shall be put under an endpoint that is using the GfJWTMiddleware.
// Reply will be of the form {"token": "TOKEN"}.
func (mw *GfJWTMiddleware) RefreshHandler(r *ghttp.Request) {
	tokenString, expire, err := mw.RefreshToken(r)
	if err != nil {
		mw.unauthorized(r, http.StatusUnauthorized, mw.HTTPStatusMessageFunc(err, r))
		return
	}

	mw.RefreshResponse(r, http.StatusOK, tokenString, expire)
}

// RefreshToken refresh token and check if token is expired
func (mw *GfJWTMiddleware) RefreshToken(r *ghttp.Request) (string, time.Time, error) {
	claims, token, err := mw.CheckIfTokenExpire(r)
	if err != nil {
		return "", time.Now(), err
	}

	// Create the token
	newToken := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))
	newClaims := newToken.Claims.(jwt.MapClaims)

	for key := range claims {
		newClaims[key] = claims[key]
	}

	expire := mw.TimeFunc().Add(mw.Timeout)
	newClaims["exp"] = expire.Unix()
	newClaims["iat"] = mw.TimeFunc().Unix()
	tokenString, err := mw.signedString(newToken)

	if err != nil {
		return "", time.Now(), err
	}

	// set cookie
	if mw.SendCookie {
		maxage := int64(expire.Unix() - time.Now().Unix())
		r.Cookie.SetCookie(mw.CookieName, tokenString, mw.CookieDomain, "/", time.Duration(maxage)*time.Second)
	}

	// set old token in blacklist
	err = mw.setBlacklist(token, claims)
	if err != nil {
		return "", time.Now(), err
	}

	return tokenString, expire, nil
}

// CheckIfTokenExpire check if token expire
func (mw *GfJWTMiddleware) CheckIfTokenExpire(r *ghttp.Request) (jwt.MapClaims, string, error) {
	token, err := mw.ParseToken(r)

	if err != nil {
		// If we receive an error, and the error is anything other than a single
		// ValidationErrorExpired, we want to return the error.
		// If the error is just ValidationErrorExpired, we want to continue, as we can still
		// refresh the token if it's within the MaxRefresh time.
		// (see https://github.com/appleboy/gin-jwt/issues/176)
		validationErr, ok := err.(*jwt.ValidationError)
		if !ok || validationErr.Errors != jwt.ValidationErrorExpired {
			return nil, "", err
		}
	}

	in, err := mw.inBlacklist(token.Raw)

	if err != nil {
		return nil, "", err
	}

	if in {
		return nil, "", ErrInvalidToken
	}

	claims := token.Claims.(jwt.MapClaims)

	origIat := int64(claims["iat"].(float64))

	if origIat < mw.TimeFunc().Add(-mw.MaxRefresh).Unix() {
		return nil, "", ErrExpiredToken
	}

	return claims, token.Raw, nil
}

// TokenGenerator method that clients can use to get a jwt token.
func (mw *GfJWTMiddleware) TokenGenerator(data interface{}) (string, time.Time, error) {
	token := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))
	claims := token.Claims.(jwt.MapClaims)

	if mw.PayloadFunc != nil {
		for key, value := range mw.PayloadFunc(data) {
			claims[key] = value
		}
	}

	expire := mw.TimeFunc().UTC().Add(mw.Timeout)
	claims["exp"] = expire.Unix()
	claims["iat"] = mw.TimeFunc().Unix()
	tokenString, err := mw.signedString(token)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expire, nil
}

func (mw *GfJWTMiddleware) jwtFromHeader(r *ghttp.Request, key string) (string, error) {
	authHeader := r.Header.Get(key)

	if authHeader == "" {
		return "", ErrEmptyAuthHeader
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == mw.TokenHeadName) {
		return "", ErrInvalidAuthHeader
	}

	return parts[1], nil
}

func (mw *GfJWTMiddleware) jwtFromQuery(r *ghttp.Request, key string) (string, error) {
	token := r.Get(key).String()

	if token == "" {
		return "", ErrEmptyQueryToken
	}

	return token, nil
}

func (mw *GfJWTMiddleware) jwtFromCookie(r *ghttp.Request, key string) (string, error) {
	cookie := r.Cookie.Get(key).String()

	if cookie == "" {
		return "", ErrEmptyCookieToken
	}

	return cookie, nil
}

func (mw *GfJWTMiddleware) jwtFromParam(r *ghttp.Request, key string) (string, error) {
	token := r.Get(key).String()
	if token == "" {
		return "", ErrEmptyParamToken
	}

	return token, nil
}

// ParseToken parse jwt token
func (mw *GfJWTMiddleware) ParseToken(r *ghttp.Request) (*jwt.Token, error) {
	var token string
	var err error

	methods := strings.Split(mw.TokenLookup, ",")
	for _, method := range methods {
		if len(token) > 0 {
			break
		}
		parts := strings.Split(strings.TrimSpace(method), ":")
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		switch k {
		case "header":
			token, err = mw.jwtFromHeader(r, v)
		case "query":
			token, err = mw.jwtFromQuery(r, v)
		case "cookie":
			token, err = mw.jwtFromCookie(r, v)
		case "param":
			token, err = mw.jwtFromParam(r, v)
		}
	}

	if err != nil {
		return nil, err
	}

	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(mw.SigningAlgorithm) != t.Method {
			return nil, ErrInvalidSigningAlgorithm
		}
		if mw.usingPublicKeyAlgo() {
			return mw.pubKey, nil
		}

		// save token string if vaild
		r.SetParam(TokenKey, token)

		return mw.Key, nil
	})
}

func (mw *GfJWTMiddleware) unauthorized(r *ghttp.Request, code int, message string) {
	r.Header.Set("WWW-Authenticate", "JWT realm="+mw.Realm)
	mw.Unauthorized(r, code, message)
	if !mw.DisabledAbort {
		r.ExitAll()
	}

}

func (mw *GfJWTMiddleware) setBlacklist(token string, claims jwt.MapClaims) error {
	// The goal of MD5 is to reduce the key length.
	token, err := gmd5.EncryptString(token)

	if err != nil {
		return err
	}

	exp := int64(claims["exp"].(float64))

	// save duration time = (exp + max_refresh) - now
	duration := time.Unix(exp, 0).Add(mw.MaxRefresh).Sub(mw.TimeFunc()).Truncate(time.Second)

	// global gcache
	err = blacklist.Set(mw.Ctx, token, true, duration)

	if err != nil {
		return err
	}

	return nil
}

func (mw *GfJWTMiddleware) inBlacklist(token string) (bool, error) {
	// The goal of MD5 is to reduce the key length.
	tokenRaw, err := gmd5.EncryptString(token)

	if err != nil {
		return false, nil
	}

	// Global gcache
	if in, err := blacklist.Contains(mw.Ctx, tokenRaw); err != nil {
		return false, nil
	} else {
		return in, nil
	}
}

// ExtractClaims help to extract the JWT claims
func ExtractClaims(r *ghttp.Request) MapClaims {
	claims := r.GetParam(PayloadKey).Interface()
	return claims.(MapClaims)
}

// GetToken help to get the JWT token string
func GetToken(r *ghttp.Request) string {
	token := r.Get(TokenKey).String()
	if len(token) == 0 {
		return ""
	}
	return token
}
